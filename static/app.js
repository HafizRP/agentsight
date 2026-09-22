// AgentSight Client Application

(function () {
  'use strict';

  // 1. Copy to Clipboard
  function setupClipboard() {
    document.addEventListener('click', function (e) {
      const copyBtn = e.target.closest('[data-copy], .copy-btn');
      if (!copyBtn) return;

      e.preventDefault();
      const textToCopy = copyBtn.getAttribute('data-copy') ||
        (copyBtn.dataset.target ? document.querySelector(copyBtn.dataset.target)?.textContent : '') ||
        copyBtn.previousElementSibling?.textContent ||
        '';

      if (!textToCopy) return;

      navigator.clipboard.writeText(textToCopy.trim()).then(function () {
        const originalHTML = copyBtn.innerHTML;
        copyBtn.innerHTML = `
          <svg class="w-4 h-4 text-emerald-400 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
          </svg>
          <span class="text-emerald-400 font-medium">Copied!</span>
        `;
        copyBtn.classList.add('border-emerald-500/50');

        setTimeout(function () {
          copyBtn.innerHTML = originalHTML;
          copyBtn.classList.remove('border-emerald-500/50');
        }, 2000);
      }).catch(function (err) {
        console.error('Failed to copy text: ', err);
      });
    });
  }

  // 2. Keyboard Shortcuts: '/' to focus search, 'Escape' to blur
  function setupKeyboardShortcuts() {
    document.addEventListener('keydown', function (e) {
      if (e.key === '/' && !['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement.tagName)) {
        e.preventDefault();
        const searchInput = document.querySelector('input[name="q"], #search-input');
        if (searchInput) {
          searchInput.focus();
          searchInput.select();
        }
      } else if (e.key === 'Escape') {
        const active = document.activeElement;
        if (active && ['INPUT', 'TEXTAREA', 'SELECT'].includes(active.tagName)) {
          active.blur();
        }
      }
    });
  }

  // 3. Smooth Number Counter Animation
  function setupAnimatedCounters() {
    const counters = document.querySelectorAll('[data-target]');
    if (!counters.length) return;

    const observer = new IntersectionObserver(function (entries, obs) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          const el = entry.target;
          const target = parseInt(el.getAttribute('data-target'), 10);
          if (isNaN(target)) return;

          const duration = 1200; // ms
          const startTime = performance.now();

          function updateCounter(currentTime) {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            // Ease-out quart
            const easeOut = 1 - Math.pow(1 - progress, 4);
            const current = Math.floor(easeOut * target);

            el.textContent = current.toLocaleString();

            if (progress < 1) {
              requestAnimationFrame(updateCounter);
            } else {
              el.textContent = target.toLocaleString();
            }
          }

          requestAnimationFrame(updateCounter);
          obs.unobserve(el);
        }
      });
    }, { threshold: 0.2 });

    counters.forEach(function (counter) {
      observer.observe(counter);
    });
  }

  // 4. Tab Switcher
  function setupTabs() {
    document.addEventListener('click', function (e) {
      const tabBtn = e.target.closest('[data-tab-target]');
      if (!tabBtn) return;

      const targetId = tabBtn.getAttribute('data-tab-target');
      const container = tabBtn.closest('[data-tabs-container]');
      if (!container) return;

      // Deactivate all tab buttons in container
      container.querySelectorAll('[data-tab-target]').forEach(function (btn) {
        btn.classList.remove('text-violet-400', 'border-violet-500', 'active-tab');
        btn.classList.add('text-slate-400', 'border-transparent');
      });

      // Activate clicked tab button
      tabBtn.classList.add('text-violet-400', 'border-violet-500', 'active-tab');
      tabBtn.classList.remove('text-slate-400', 'border-transparent');

      // Hide all panels
      container.querySelectorAll('[data-tab-panel]').forEach(function (panel) {
        panel.classList.add('hidden');
      });

      // Show targeted panel
      const targetPanel = container.querySelector(targetId);
      if (targetPanel) {
        targetPanel.classList.remove('hidden');
      }
    });
  }

  // 5. HTMX Hooks
  function setupHTMXHooks() {
    if (typeof document.body.addEventListener !== 'function') return;

    // Handle 401 Unauthorized globally for HTMX requests
    document.body.addEventListener('htmx:responseError', function (evt) {
      if (evt.detail && evt.detail.xhr && evt.detail.xhr.status === 401) {
        window.location.href = '/login';
      }
    });

    // Re-run code highlighting after HTMX swaps
    document.body.addEventListener('htmx:afterSwap', function (evt) {
      if (window.hljs) {
        evt.detail.target.querySelectorAll('pre code').forEach(function (block) {
          window.hljs.highlightElement(block);
        });
      }
    });
  }

  // Initialize all features on DOM ready
  document.addEventListener('DOMContentLoaded', function () {
    setupClipboard();
    setupKeyboardShortcuts();
    setupAnimatedCounters();
    setupTabs();
    setupHTMXHooks();

    // Syntax highlighting initial pass
    if (window.hljs) {
      document.querySelectorAll('pre code').forEach(function (block) {
        window.hljs.highlightElement(block);
      });
    }
  });
})();
