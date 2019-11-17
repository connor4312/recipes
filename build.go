package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/pkg/errors"
)

// Page is the template rendered into page.html.
type Page struct {
	Title    string
	Contents template.HTML
}

// RecipeRaw is the structure corresponding to each recipe.json.
type RecipeRaw struct {
	Name        string
	Rating      float32
	Notes       []string
	Source      string
	Ingredients []struct{ Ingredient, Quantity string }
	Steps       []string
}

// Recipe a wrapped RecipeRaw containing some additional generated metadata.
type Recipe struct {
	RecipeRaw
	Slug      string
	ImagePath string
}

// PageName returns the page of the recipe's HTML file.
func (r Recipe) PageName() string { return r.Slug + ".html" }

// ImageName returns the location of the recipe's image file.
func (r Recipe) ImageName() string { return r.Slug + ".jpg" }

const (
	outDir = "dist"
)

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"getOrigin": func(rawURL string) template.URL {
		u, err := url.Parse(rawURL)
		if err != nil {
			return template.URL(rawURL)
		}

		return template.URL(u.Host)
	},

	"renderMd": func(md string) template.HTML {
		return template.HTML(markdown.ToHTML([]byte(md), nil, nil))
	},
}).ParseFiles(
	filepath.Join("support", "home.html"),
	filepath.Join("support", "page.html"),
	filepath.Join("support", "recipe.html"),
))

func main() {
	start := time.Now()
	if err := mkdir(outDir); err != nil {
		fmt.Printf("error making dist folder: %s", err)
		os.Exit(1)
	}

	recipes, err := gatherRecipes()
	if err != nil {
		fmt.Printf("error reading recipes: %s", err)
		os.Exit(1)
	}

	fmt.Printf("discovered %d recipes\r\n", len(recipes))

	var wg sync.WaitGroup
	wg.Add(len(recipes))

	go func() {
		if err := writeHome(recipes); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		wg.Done()
	}()

	for _, r := range recipes {
		go func(r *Recipe) {
			if err := writeRecipe(r); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			wg.Done()
		}(r)
	}

	wg.Wait()

	fmt.Printf("recipe generation complete in %dms\r\n", time.Now().Sub(start)/time.Millisecond)
}

func gatherRecipes() (recipes []*Recipe, err error) {
	children, err := ioutil.ReadDir("recipes")
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		if !child.IsDir() {
			continue
		}

		path := filepath.Join("recipes", child.Name(), "recipe.json")
		if !fileExists(path) {
			continue
		}

		recipe := &Recipe{Slug: child.Name()}
		if err := readJSONFile(path, &recipe.RecipeRaw); err != nil {
			return nil, err
		}

		if image := filepath.Join("recipes", child.Name(), "image.jpg"); fileExists(image) {
			recipe.ImagePath = image
		}

		recipes = append(recipes, recipe)
	}

	sort.Slice(recipes, func(i, j int) bool { return recipes[i].Rating < recipes[j].Rating })

	return recipes, nil
}

func writeHome(recipes []*Recipe) error {
	readme, err := ioutil.ReadFile("readme.md")
	if err != nil {
		return errors.Wrap(err, "error opening readme")
	}

	homeData := struct {
		Recipes []*Recipe
		Readme  string
	}{
		Recipes: recipes,
		Readme:  string(readme),
	}

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "home.html", homeData); err != nil {
		return errors.Wrap(err, "error rendering homepage")
	}

	return writePageToDisk("index.html", Page{Title: "Recipes", Contents: template.HTML(buf.String())})
}

func writeRecipe(recipe *Recipe) error {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "recipe.html", recipe); err != nil {
		return errors.Wrapf(err, "error rendering recipe %s", recipe.PageName())
	}

	page := Page{
		Title:    recipe.Name,
		Contents: template.HTML(buf.String()),
	}

	if err := writePageToDisk(recipe.PageName(), page); err != nil {
		return err
	}

	if recipe.ImagePath != "" {
		if err := copyFile(filepath.Join("dist", recipe.ImageName()), recipe.ImagePath); err != nil {
			return err
		}
	}

	return nil
}

func writePageToDisk(name string, data Page) error {
	f, err := os.Create(filepath.Join(outDir, name))
	if err != nil {
		return errors.Wrapf(err, "error creating file %s", name)
	}
	defer f.Close()

	if err := templates.ExecuteTemplate(f, "page.html", data); err != nil {
		return errors.Wrapf(err, "error outer page for %s", name)
	}

	return nil
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func mkdir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.Mkdir(dir, os.ModePerm)
	}

	return err
}

func readJSONFile(file string, out interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "error reading %s", file)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(out); err != nil {
		return errors.Wrapf(err, "error reading %s", file)
	}

	return nil
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
