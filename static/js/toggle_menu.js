(() => {
  const button = document.querySelector('.burger-button');
  const menuContent = document.querySelector('.mobile-menu-content');
  const overlay = document.querySelector('#sidebar .overlay');

  const toggleMenu = () => {
    menuContent.classList.toggle('menu-active');

    if (overlay.style.display === 'block') {
      overlay.style.display = 'none';

      return;
    }

    overlay.style.display = 'block';
  }

  button.addEventListener('click', (e) => {
    if (e.target === button) {
      toggleMenu();
    }
  });

  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) {
      toggleMenu();
    }
  })

  menuContent.addEventListener('click', (e) => {
    if (e.target.tagName === 'A') {
      toggleMenu();
    }
  })
})();
