// NY Times:
copy({
  source: window.location.origin + window.location.pathname,
  ingredients: [...document.querySelectorAll('.recipe-ingredients li')].map(el => ({
    ingredient: el.querySelector('.ingredient-name').innerText,
    quantity: el.querySelector('.quantity').innerText,
  })),
  steps: [...document.querySelectorAll('.recipe-steps li')].map(el => el.innerText),
});

// Common wordpress recipe system:
copy({
  source: window.location.origin + window.location.pathname,
  ingredients: [...document.querySelectorAll('.wprm-recipe-ingredient')].map(el => {
    const amt = el.querySelector('.wprm-recipe-ingredient-amount');
    const unit = el.querySelector('.wprm-recipe-ingredient-unit');
    return {
      quantity: ((amt ? amt.innerText : '') + ' ' + (unit ? unit.innerText : '')).trim(),
      ingredient: el.querySelector('.wprm-recipe-ingredient-name').innerText,
    };
  }),
  steps: [...document.querySelectorAll('.wprm-recipe-instruction-text')].map(el => el.innerText),
});
