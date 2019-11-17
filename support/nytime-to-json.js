copy({
  source: window.location.origin + window.location.pathname,
  ingredients: [...document.querySelectorAll('.recipe-ingredients li')].map(el => ({
    ingredient: el.querySelector('.ingredient-name').innerText,
    quantity: el.querySelector('.quantity').innerText,
  })),
  steps: [...document.querySelectorAll('.recipe-steps li')].map(el => el.innerText),
});
