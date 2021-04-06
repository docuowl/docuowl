import init, { setup, search } from './owl_wasm.js';
const debounce = (n,t,u) => {var e;return function(){var i=this,o=arguments,a=u&&!e;clearTimeout(e),e=setTimeout(function(){e=null,u||n.apply(i,o)},t),a&&n.apply(i,o)}};

(() => {
    async function run() {
        await init();
        const ftsData = document.querySelector('meta[name="owl-fts-index"]').getAttribute('content');
        const instance = setup(ftsData);
        console.log('Initialisation done');

        const searchButton = document.querySelector('div.search');
        const body = document.querySelector('body');
        const searchInput = document.querySelector('#search-input');
        const searchResultList = document.querySelector('#result-list');
        const searchContainer = document.querySelector('#search-container');
        let searchResults = [];
        let searchActive = false;
        let activeIndex = -1;

        const showSearch = () => {
            body.classList.add('showing-search');
            searchInput.focus();
            searchActive = true;
        };

        const hideSearch = () => {
            body.classList.remove('showing-search');
            searchActive = false;
            activeIndex = -1;
            searchResults = [];
            searchResultList.innerHTML = '';
            searchInput.value = '';
            searchInput.blur();
        };

        const changeActiveIndex = (newIndex) => {
            activeIndex = newIndex;
            let idx = 0;
            for (let el of searchResultList.children) {
                if (idx === activeIndex) {
                    el.classList.add('active');
                } else {
                    el.classList.remove('active');
                }
                idx++;
            }
        }

        const doSearch = debounce((terms) => {
            terms = terms.trim();
            if (terms.length === 0) {
                return;
            }
            const then = Date.now();
            searchResults = search(instance, terms);
            const now = Date.now();
            console.log(`[owl-fts] Search took ${now - then}ms`);
            searchResultList.innerHTML = '';
            activeIndex = -1;
            if (searchResults.length == 0) {
                searchResultList.innerHTML = `<li class="empty">No results for ${terms}</li>`;
                return;
            }

            let idx = 0;
            const html = searchResults.map(r => {
                let anchor = document.querySelector(`#anchor-${r.anchor}`);
                return `<li class="result" data-to="content-${r.anchor}" data-idx="${idx++}">${anchor.textContent}</li>`;
            }).join('');

            searchResultList.innerHTML = html;

            for (let el of searchResultList.querySelectorAll('.result')) {
                el.addEventListener('mouseenter', (e) => {
                    changeActiveIndex(parseInt(e.target.getAttribute('data-idx'), 10));
                });
                el.addEventListener('click', (e) => {
                    let el = e.target;
                    el.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                    document.location.hash = el.getAttribute('data-to');
                    hideSearch();
                })
            }
        }, 300);

        searchContainer.addEventListener('click', (e) => {
            if (e.target === searchContainer) {
                hideSearch();
            }
        });

        window.addEventListener('keydown', (e) => {
            if (((e.ctrlKey || e.metaKey) && e.key == 'f') || (e.key == '/' && !searchActive)) {
                e.preventDefault();
                showSearch();
                return;
            }

            if (e.key == 'Escape' && searchActive) {
                e.preventDefault();
                hideSearch();
                return;
            }

            if (e.key == 'ArrowUp' && searchActive) {
                e.preventDefault();
                e.stopPropagation();
                if (activeIndex > 0) {
                    changeActiveIndex(activeIndex - 1);
                }
            }

            if (e.key == 'ArrowDown' && searchActive) {
                e.preventDefault();
                e.stopPropagation();
                if (activeIndex + 1 < searchResults.length) {
                    changeActiveIndex(activeIndex + 1);
                }
            }
        });

        searchInput.addEventListener('keyup', (e) => {
            if ((e.key == 'ArrowDown' || e.key == 'ArrowUp') && searchResults.length != 0) {
                e.preventDefault();
                e.stopPropagation();
                return;
            }

           if (e.key == 'Enter') {
                e.preventDefault();
                e.stopPropagation();
                if (activeIndex != -1) {
                    let el = searchResultList.querySelector('.active');
                    el.click();
                }
               return;
           }

           doSearch(searchInput.value);
        });

        searchButton.onclick = (e) => {
            e.preventDefault();
            e.stopPropagation();
            showSearch();
        };
    };

    document.addEventListener('DOMContentLoaded', () => {
        run();
    });
})();
