(() => {
  const currentTheme = localStorage.getItem('theme') || 'dark';
  const selector = document.querySelector('#theme-selector');
  const body = document.body;
  selector.checked = currentTheme === 'dark';
  if (currentTheme === 'light') {
    body.classList.add('light');
  }

  selector.addEventListener('change', () => {
    localStorage.setItem('theme', selector.checked ? 'dark' : 'light');
    if (selector.checked) {
      body.classList.remove('light');
    } else {
      body.classList.add('light');
    }
  });
})();
