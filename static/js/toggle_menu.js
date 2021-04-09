(() => {
  const button = document.querySelector('.burger-button');
  const menuContent = document.querySelector('#mobile-menu-content');

  const toggleMenu = () => {
    if (menuContent.style.display === 'block') {
      menuContent.style.display = 'none';
      return;
    }
    menuContent.style.display = 'block';
  }

  button.addEventListener('click', (e) => {
    if (e.target === button) {
      toggleMenu();
    }
  });
})();
