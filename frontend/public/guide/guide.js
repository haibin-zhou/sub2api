(function () {
  'use strict';

  var AUTH_TOKEN_KEY = 'auth_token';
  var AUTH_USER_KEY = 'auth_user';
  var REFRESH_TOKEN_KEY = 'refresh_token';
  var TOKEN_EXPIRES_AT_KEY = 'token_expires_at';
  var TOKEN_REFRESH_LOCK_NAME = 'sub2api-auth-token-refresh';
  var LOGIN_REDIRECT_URL =
    '/login?redirect=' + encodeURIComponent('/guide.html');

  // 0. Sync light/dark theme with main sub2api console (`localStorage.theme`)
  var themeToggle = document.getElementById('theme-toggle');

  function applyTheme(isDark) {
    document.documentElement.classList.toggle('dark', isDark);
    if (themeToggle) {
      themeToggle.textContent = isDark ? '☀' : '☾';
      themeToggle.setAttribute(
        'aria-label',
        isDark ? '切换到浅色模式' : '切换到深色模式'
      );
      themeToggle.setAttribute(
        'title',
        isDark ? '切换到浅色模式' : '切换到深色模式'
      );
    }
  }

  (function initTheme() {
    var saved = null;
    try {
      saved = window.localStorage.getItem('theme');
    } catch (_err) {
      saved = null;
    }
    var prefersDark =
      window.matchMedia &&
      window.matchMedia('(prefers-color-scheme: dark)').matches;
    applyTheme(saved ? saved === 'dark' : prefersDark);
  })();

  if (themeToggle) {
    themeToggle.addEventListener('click', function () {
      var nextDark = !document.documentElement.classList.contains('dark');
      applyTheme(nextDark);
      try {
        window.localStorage.setItem('theme', nextDark ? 'dark' : 'light');
      } catch (_err) {
        // Ignore storage errors in private/file contexts
      }
    });
  }

  // 1. Auth gate: require valid login session on http(s), redirect to /login otherwise
  function safeGetItem(key) {
    try {
      return window.localStorage.getItem(key);
    } catch (_err) {
      return null;
    }
  }

  function clearAuthStorage() {
    try {
      window.localStorage.removeItem(AUTH_TOKEN_KEY);
      window.localStorage.removeItem(REFRESH_TOKEN_KEY);
      window.localStorage.removeItem(AUTH_USER_KEY);
      window.localStorage.removeItem(TOKEN_EXPIRES_AT_KEY);
    } catch (_err) {
      // Ignore storage errors
    }
  }

  function redirectToLogin(shouldClear) {
    if (shouldClear) clearAuthStorage();
    document.body.classList.add('auth-pending');
    window.location.replace(LOGIN_REDIRECT_URL);
  }

  function unlockPage() {
    document.body.classList.remove('auth-pending');
    var gate = document.getElementById('auth-gate');
    if (gate) gate.setAttribute('aria-hidden', 'true');
  }

  function fetchCurrentUser(token) {
    return window
      .fetch('/api/v1/auth/me', {
        method: 'GET',
        cache: 'no-store',
        headers: {
          Authorization: 'Bearer ' + token,
          Accept: 'application/json'
        }
      })
      .then(function (res) {
        if (!res.ok) {
          return { ok: false, status: res.status };
        }
        return res
          .json()
          .then(function (payload) {
            var valid =
              payload &&
              payload.code === 0 &&
              payload.data &&
              typeof payload.data.id !== 'undefined';
            return { ok: Boolean(valid), status: res.status };
          })
          .catch(function () {
            return { ok: false, status: res.status };
          });
      })
      .catch(function () {
        return { ok: false, status: 0 };
      });
  }

  function performTokenRefresh(failedToken) {
    var runner = function () {
      // If another console tab already rotated the token, adopt it without re-refreshing
      var latestToken = safeGetItem(AUTH_TOKEN_KEY);
      if (latestToken && latestToken !== failedToken) {
        return Promise.resolve(latestToken);
      }

      var refreshToken = safeGetItem(REFRESH_TOKEN_KEY);
      if (!refreshToken) {
        return Promise.resolve(null);
      }

      return window
        .fetch('/api/v1/auth/refresh', {
          method: 'POST',
          cache: 'no-store',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json'
          },
          body: JSON.stringify({ refresh_token: refreshToken })
        })
        .then(function (res) {
          if (!res.ok) return null;
          return res.json().catch(function () {
            return null;
          });
        })
        .then(function (payload) {
          if (
            !payload ||
            payload.code !== 0 ||
            !payload.data ||
            !payload.data.access_token
          ) {
            return null;
          }
          try {
            window.localStorage.setItem(
              AUTH_TOKEN_KEY,
              payload.data.access_token
            );
            if (payload.data.expires_in) {
              window.localStorage.setItem(
                TOKEN_EXPIRES_AT_KEY,
                String(Date.now() + payload.data.expires_in * 1000)
              );
            }
            if (payload.data.refresh_token) {
              // Write rotating refresh token last as commit marker (matches tokenRefresh.ts)
              window.localStorage.setItem(
                REFRESH_TOKEN_KEY,
                payload.data.refresh_token
              );
            }
          } catch (_err) {
            return null;
          }
          return payload.data.access_token;
        })
        .catch(function () {
          return null;
        });
    };

    if (navigator.locks && typeof navigator.locks.request === 'function') {
      return navigator.locks.request(TOKEN_REFRESH_LOCK_NAME, runner);
    }
    return runner();
  }

  function verifyAuthAndInit() {
    // Allow local offline preview when opened directly via file://
    if (window.location.protocol === 'file:') {
      unlockPage();
      initGuideInteractions();
      return;
    }

    var token = safeGetItem(AUTH_TOKEN_KEY);
    if (!token) {
      redirectToLogin(false);
      return;
    }

    fetchCurrentUser(token).then(function (result) {
      if (result.ok) {
        unlockPage();
        initGuideInteractions();
        watchCrossTabLogout();
        return;
      }

      if (result.status === 401 && safeGetItem(REFRESH_TOKEN_KEY)) {
        performTokenRefresh(token).then(function (newToken) {
          if (!newToken) {
            redirectToLogin(true);
            return;
          }
          fetchCurrentUser(newToken).then(function (retryResult) {
            if (retryResult.ok) {
              unlockPage();
              initGuideInteractions();
              watchCrossTabLogout();
            } else {
              redirectToLogin(retryResult.status === 401);
            }
          });
        });
        return;
      }

      redirectToLogin(result.status === 401 || result.status === 403);
    });
  }

  function watchCrossTabLogout() {
    window.addEventListener('storage', function (event) {
      if (event.key === AUTH_TOKEN_KEY && !event.newValue) {
        redirectToLogin(false);
      }
    });
  }

  // 2. Guide interactive widgets (copy buttons, lightbox, scrollspy, back-to-top)
  function initGuideInteractions() {
    function fallbackCopy(text) {
      var textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.setAttribute('readonly', '');
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      var ok = false;
      try {
        ok = document.execCommand('copy');
      } catch (_err) {
        ok = false;
      }
      document.body.removeChild(textarea);
      return ok;
    }

    function showCopiedState(button) {
      var original = button.getAttribute('data-label') || button.textContent;
      button.setAttribute('data-label', original);
      button.textContent = '已复制 ✓';
      button.classList.add('is-copied');
      window.clearTimeout(button._copyTimer);
      button._copyTimer = window.setTimeout(function () {
        button.textContent = original;
        button.classList.remove('is-copied');
      }, 1800);
    }

    document.querySelectorAll('.copy-btn[data-copy]').forEach(function (button) {
      button.addEventListener('click', function () {
        var text = button.getAttribute('data-copy') || '';
        if (!text) return;
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard
            .writeText(text)
            .then(function () {
              showCopiedState(button);
            })
            .catch(function () {
              if (fallbackCopy(text)) showCopiedState(button);
            });
        } else if (fallbackCopy(text)) {
          showCopiedState(button);
        }
      });
    });

    // Click-to-zoom screenshot lightbox
    var figureButtons = Array.prototype.slice.call(
      document.querySelectorAll('.figure-btn')
    );
    var lightbox = document.getElementById('lightbox');
    var lightboxImg = document.getElementById('lightbox-img');
    var lightboxCaption = document.getElementById('lightbox-caption');
    var lightboxClose = document.getElementById('lightbox-close');
    var lightboxPrev = document.getElementById('lightbox-prev');
    var lightboxNext = document.getElementById('lightbox-next');
    var currentIndex = -1;

    function renderLightbox(index) {
      if (!lightbox || !lightboxImg || figureButtons.length === 0) return;
      currentIndex = (index + figureButtons.length) % figureButtons.length;
      var btn = figureButtons[currentIndex];
      var img = btn.querySelector('img');
      var figure = btn.closest('figure');
      var captionEl = figure ? figure.querySelector('.caption-text') : null;
      if (!img) return;
      lightboxImg.src = img.currentSrc || img.src;
      lightboxImg.alt = img.alt || '';
      if (lightboxCaption) {
        var label = captionEl ? captionEl.textContent : img.alt || '';
        lightboxCaption.textContent =
          label +
          ' (' +
          (currentIndex + 1) +
          ' / ' +
          figureButtons.length +
          ')';
      }
    }

    function openLightbox(index) {
      if (!lightbox) return;
      renderLightbox(index);
      lightbox.classList.add('is-open');
      lightbox.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
    }

    function closeLightbox() {
      if (!lightbox) return;
      lightbox.classList.remove('is-open');
      lightbox.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
      currentIndex = -1;
    }

    figureButtons.forEach(function (btn, idx) {
      btn.addEventListener('click', function () {
        openLightbox(idx);
      });
    });

    if (lightboxClose) {
      lightboxClose.addEventListener('click', closeLightbox);
    }
    if (lightboxPrev) {
      lightboxPrev.addEventListener('click', function (e) {
        e.stopPropagation();
        renderLightbox(currentIndex - 1);
      });
    }
    if (lightboxNext) {
      lightboxNext.addEventListener('click', function (e) {
        e.stopPropagation();
        renderLightbox(currentIndex + 1);
      });
    }
    if (lightbox) {
      lightbox.addEventListener('click', function (e) {
        if (e.target === lightbox) closeLightbox();
      });
    }

    document.addEventListener('keydown', function (e) {
      if (!lightbox || !lightbox.classList.contains('is-open')) return;
      if (e.key === 'Escape') {
        closeLightbox();
      } else if (e.key === 'ArrowLeft') {
        renderLightbox(currentIndex - 1);
      } else if (e.key === 'ArrowRight') {
        renderLightbox(currentIndex + 1);
      }
    });

    // TOC scrollspy + mobile TOC auto-close
    var tocLinks = Array.prototype.slice.call(
      document.querySelectorAll('.toc .toc-link[href^="#"]')
    );
    var mobileDetails = document.querySelector('.mobile-toc');
    document
      .querySelectorAll('.mobile-toc .toc-link[href^="#"]')
      .forEach(function (link) {
        link.addEventListener('click', function () {
          if (mobileDetails) mobileDetails.removeAttribute('open');
        });
      });

    if ('IntersectionObserver' in window && tocLinks.length > 0) {
      var linkById = {};
      var targets = [];
      tocLinks.forEach(function (link) {
        var id = link.getAttribute('href').slice(1);
        var target = document.getElementById(id);
        if (target) {
          linkById[id] = link;
          targets.push(target);
        }
      });

      function setActive(id) {
        tocLinks.forEach(function (link) {
          link.classList.remove('is-active');
        });
        if (linkById[id]) {
          linkById[id].classList.add('is-active');
        }
      }

      var observer = new IntersectionObserver(
        function (entries) {
          entries.forEach(function (entry) {
            if (entry.isIntersecting) {
              setActive(entry.target.id);
            }
          });
        },
        { rootMargin: '-90px 0px -65% 0px', threshold: 0.01 }
      );

      targets.forEach(function (el) {
        observer.observe(el);
      });
      if (targets[0]) setActive(targets[0].id);
    }

    // Back to top button
    var backToTop = document.getElementById('back-to-top');
    if (backToTop) {
      function toggleBackToTop() {
        if (window.scrollY > 360) {
          backToTop.classList.add('is-visible');
        } else {
          backToTop.classList.remove('is-visible');
        }
      }
      window.addEventListener('scroll', toggleBackToTop, { passive: true });
      toggleBackToTop();
      backToTop.addEventListener('click', function () {
        window.scrollTo({ top: 0, behavior: 'smooth' });
      });
    }
  }

  verifyAuthAndInit();
})();
