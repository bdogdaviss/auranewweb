/* =============================================
   AURA OPTIMIZER — Main JavaScript
   ============================================= */

(function () {
  'use strict';

  /* -----------------------------------------
     TAB / NAVIGATION SYSTEM
     ----------------------------------------- */
  const navLinks = document.querySelectorAll('[data-section]');
  const tabs = document.querySelectorAll('.tab-content');

  function switchTab(sectionName) {
    // Deactivate all tabs
    tabs.forEach(tab => {
      tab.classList.remove('active');
    });

    // Deactivate all nav links
    document.querySelectorAll('.nav-link').forEach(link => {
      link.classList.remove('active');
    });

    // Activate target tab
    const targetTab = document.getElementById('tab-' + sectionName);
    if (targetTab) {
      targetTab.classList.add('active');
    }

    // Activate matching nav links
    document.querySelectorAll('.nav-link[data-section="' + sectionName + '"]').forEach(link => {
      link.classList.add('active');
    });

    // Scroll to top smoothly
    window.scrollTo({ top: 0, behavior: 'smooth' });

    // Trigger staggered card animations after a short delay
    setTimeout(() => {
      triggerCardAnimations(targetTab);
    }, 100);
  }

  function triggerCardAnimations(container) {
    if (!container) return;
    const cards = container.querySelectorAll('.scroll-animate-card, .scroll-animate');
    cards.forEach((card, i) => {
      card.classList.remove('visible');
      card.style.transitionDelay = (i * 100) + 'ms';
      // Force reflow
      void card.offsetWidth;
      setTimeout(() => {
        card.classList.add('visible');
      }, 50);
    });
  }

  // Bind click events
  navLinks.forEach(link => {
    link.addEventListener('click', function (e) {
      e.preventDefault();
      const section = this.getAttribute('data-section');
      if (section) {
        switchTab(section);
      }
    });
  });

  /* -----------------------------------------
     HEADER SCROLL BEHAVIOR
     ----------------------------------------- */
  const header = document.getElementById('header');
  let lastScrollY = 0;

  function handleScroll() {
    const currentScrollY = window.scrollY;

    // Scrolled class
    if (currentScrollY > 50) {
      header.classList.add('scrolled');
    } else {
      header.classList.remove('scrolled');
    }

    // Collapsed class (only on desktop)
    if (window.innerWidth > 1024) {
      if (currentScrollY > 150 && currentScrollY > lastScrollY) {
        header.classList.add('collapsed');
      } else {
        header.classList.remove('collapsed');
      }
    }

    lastScrollY = currentScrollY;
  }

  window.addEventListener('scroll', handleScroll, { passive: true });

  // Un-collapse on hover
  header.addEventListener('mouseenter', function () {
    header.classList.remove('collapsed');
  });

  /* -----------------------------------------
     CAROUSEL SYSTEM
     ----------------------------------------- */
  const carouselState = {};

  function carouselInit(id, totalSlides) {
    carouselState[id] = {
      current: 0,
      total: totalSlides
    };
  }

  window.carouselGoTo = function (id, index) {
    const state = carouselState[id];
    if (!state) return;

    // Wrap index
    if (index < 0) index = state.total - 1;
    if (index >= state.total) index = 0;

    state.current = index;

    const track = document.getElementById('carousel-track-' + id);
    if (track) {
      track.style.transform = 'translateX(-' + (index * 100) + '%)';
    }

    // Update dots
    const dotsContainer = document.getElementById('carousel-dots-' + id);
    if (dotsContainer) {
      const dots = dotsContainer.querySelectorAll('.carousel-dot');
      dots.forEach((dot, i) => {
        if (i === index) {
          dot.classList.add('active');
        } else {
          dot.classList.remove('active');
        }
      });
    }
  };

  window.carouselMove = function (id, dir) {
    const state = carouselState[id];
    if (!state) return;
    window.carouselGoTo(id, state.current + dir);
  };

  // Initialize bundle carousel
  carouselInit('bundle', 3);

  /* -----------------------------------------
     INTERSECTION OBSERVER (Scroll Animations)
     ----------------------------------------- */
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px 80px 0px'
  };

  // Observer for section headers
  const headerObserver = new IntersectionObserver(function (entries) {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        headerObserver.unobserve(entry.target);
      }
    });
  }, observerOptions);

  // Observer for cards (staggered)
  const cardObserver = new IntersectionObserver(function (entries) {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        // Find sibling cards for stagger
        const parent = entry.target.parentElement;
        if (parent) {
          const siblings = parent.querySelectorAll('.scroll-animate-card');
          siblings.forEach((card, i) => {
            card.style.transitionDelay = (i * 120) + 'ms';
            card.classList.add('visible');
          });
        }
        cardObserver.unobserve(entry.target);
      }
    });
  }, observerOptions);

  function initObservers() {
    // Observe section headers
    document.querySelectorAll('.scroll-animate').forEach(el => {
      headerObserver.observe(el);
    });

    // Observe first card in each group to trigger stagger
    const cardGroups = new Set();
    document.querySelectorAll('.scroll-animate-card').forEach(card => {
      const parent = card.parentElement;
      if (parent && !cardGroups.has(parent)) {
        cardGroups.add(parent);
        cardObserver.observe(card);
      }
    });
  }

  initObservers();

  /* -----------------------------------------
     ANTI-INSPECT
     ----------------------------------------- */
  document.addEventListener('contextmenu', function (e) {
    e.preventDefault();
  });

  document.addEventListener('keydown', function (e) {
    if (e.key === 'F12' ||
      (e.ctrlKey && e.shiftKey && (e.key === 'I' || e.key === 'J')) ||
      (e.ctrlKey && e.key === 'u')) {
      e.preventDefault();
    }
  });

  /* -----------------------------------------
     INITIAL PAGE SETUP
     ----------------------------------------- */
  // Trigger animations for the initially active tab (Home)
  const activeTab = document.querySelector('.tab-content.active');
  if (activeTab) {
    setTimeout(() => {
      triggerCardAnimations(activeTab);
    }, 800);
  }

})();
