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

        const findCorrelations = (terms, result) => {
            const containerText = document
                .querySelector(`#content-${result.anchor} .content`)
                .textContent
                .toLowerCase()
                .replace(/\n/g, ' ')
                .replace(/\s+/g, ' ')
                .trim();

            const allIndexes = terms
                .map(t => containerText.includes(t) ? { 
                    start: Math.max(0, idx - 100),
                    end: idx + 100
                } : false)
                .filter(Boolean);

            if (allIndexes.length === 0) {
                return null;
            }

            const filteredIndexes = [];

            for (let idx of allIndexes) {
                if (filteredIndexes.length >= 2) {
                    break;
                }
                if (filteredIndexes.length === 0) {
                    filteredIndexes.push(idx);
                    continue;
                }
                if (!filteredIndexes.some(i => i.start <= idx.start && i.end >= idx.end)) {
                    filteredIndexes.push(idx);
                }
            }

            let resultingText = filteredIndexes.map(i => {
                let txt = containerText.substring(i.start, i.end);
                if (i.start !== 0) {
                    txt = '&hellip;' + txt.substring(txt.indexOf(' '));
                }
                return txt + '&hellip;';
            }).join(' ');

            terms.forEach(t => {
                resultingText = resultingText.replace(t, (x) => {
                    return `<span class="search-result-match">${x}</span>`;
                })
            });

            return resultingText;
        };

        const doSearch = debounce((terms) => {
            terms = terms.trim().toLowerCase();
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

            const allTerms = terms.split(' ').map(i => i.trim()).filter(Boolean);
            const html = searchResults.map(r => {
                const correlations = findCorrelations(allTerms, r);
                if (correlations === null) {
                    return null;
                }
                const anchor = document.querySelector(`#anchor-${r.anchor}`);
                return `<li class="result" data-to="content-${r.anchor}" data-idx="${idx++}"><div class="search-title">${anchor.textContent}</div><div class="text-matches">${correlations}</div></li>`;
            }).filter(Boolean).join('');

            searchResultList.innerHTML = html;

            for (let el of searchResultList.querySelectorAll('.result')) {
                el.addEventListener('mouseenter', (e) => {
                    changeActiveIndex(parseInt(e.target.getAttribute('data-idx'), 10));
                });
                el.addEventListener('click', (e) => {
                    let obj = e.target;
                    while (obj && !obj.classList.contains('result')) {
                        obj = obj.parentNode;
                    }
                    let el = document.querySelector(`#${obj.getAttribute('data-to')}`);
                    el.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                    document.location.hash = obj.getAttribute('data-to');
                    hideSearch();
                });
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
