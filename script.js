// ─── CONFIGURATION ────────────────────────────────────────────
// Détection automatique de l'URL de l'API (même origine que le site)
const API_BASE_URL = window.location.origin;

const GAMES = [
  { id: 32440,   title: "LEGO® Star Wars™",          subtitle: "The Complete Saga",         series: "Star Wars",      tag: "action",   appid: 32440  },
  { id: 32450,   title: "LEGO® Star Wars™ III",       subtitle: "The Clone Wars",            series: "Star Wars",      tag: "action",   appid: 32450  },
  { id: 432200,  title: "LEGO® Star Wars™",           subtitle: "The Force Awakens",         series: "Star Wars",      tag: "action",   appid: 432200 },
  { id: 920210,  title: "LEGO® Star Wars™",           subtitle: "The Skywalker Saga",        series: "Star Wars",      tag: "action",   appid: 920210 },
  { id: 32400,   title: "LEGO® Indiana Jones™",       subtitle: "The Original Adventures",   series: "Indiana Jones",  tag: "aventure", appid: 32400  },
  { id: 32460,   title: "LEGO® Indiana Jones™ 2",     subtitle: "The Adventure Continues",   series: "Indiana Jones",  tag: "aventure", appid: 32460  },
  { id: 428740,  title: "LEGO® Harry Potter™",        subtitle: "Years 1-4",                 series: "Harry Potter",   tag: "magie",    appid: 428740 },
  { id: 428750,  title: "LEGO® Harry Potter™",        subtitle: "Years 5-7",                 series: "Harry Potter",   tag: "magie",    appid: 428750 },
  { id: 203160,  title: "LEGO® The Lord of the Rings™", subtitle: "",                        series: "Middle-Earth",   tag: "aventure", appid: 203160 },
  { id: 261570,  title: "LEGO® The Hobbit™",          subtitle: "",                          series: "Middle-Earth",   tag: "aventure", appid: 261570 },
  { id: 249130,  title: "LEGO® Marvel™ Super Heroes", subtitle: "",                          series: "Marvel",         tag: "action",   appid: 249130 },
  { id: 647830,  title: "LEGO® Marvel™ Super Heroes 2", subtitle: "",                        series: "Marvel",         tag: "action",   appid: 647830 },
  { id: 438100,  title: "LEGO® Marvel's Avengers",    subtitle: "",                          series: "Marvel",         tag: "action",   appid: 438100 },
  { id: 608790,  title: "LEGO® DC Super-Villains",    subtitle: "",                          series: "DC",             tag: "action",   appid: 608790 },
  { id: 21000,   title: "LEGO® Batman™",              subtitle: "The Videogame",             series: "Batman",         tag: "action",   appid: 21000  },
  { id: 204060,  title: "LEGO® Batman™ 2",            subtitle: "DC Super Heroes",           series: "Batman",         tag: "action",   appid: 204060 },
  { id: 313690,  title: "LEGO® Batman™ 3",            subtitle: "Beyond Gotham",             series: "Batman",         tag: "action",   appid: 313690 },
  { id: 352400,  title: "LEGO® Jurassic World",       subtitle: "",                          series: "Jurassic",       tag: "aventure", appid: 352400 },
  { id: 353090,  title: "LEGO® City Undercover",      subtitle: "",                          series: "City",           tag: "sandbox",  appid: 353090 },
  { id: 267530,  title: "The LEGO® Movie",            subtitle: "Videogame",                 series: "LEGO Movie",     tag: "aventure", appid: 267530 },
  { id: 787480,  title: "The LEGO® Movie 2",          subtitle: "Videogame",                 series: "LEGO Movie",     tag: "aventure", appid: 787480 },
  { id: 332310,  title: "LEGO® Worlds",               subtitle: "",                          series: "Worlds",         tag: "sandbox",  appid: 332310 },
  { id: 1386430, title: "LEGO® Brawls",               subtitle: "",                          series: "Brawls",         tag: "combat",   appid: 1386430 },
  { id: 1429400, title: "LEGO® Builder's Journey",    subtitle: "",                          series: "Builder",        tag: "puzzle",   appid: 1429400 },
  { id: 1898790, title: "LEGO® Bricktales",           subtitle: "",                          series: "Bricktales",     tag: "puzzle",   appid: 1898790 },
  { id: 2677660, title: "LEGO® Horizon Adventures",   subtitle: "",                          series: "Horizon",        tag: "action",   appid: 2677660 },
  { id: 0,       title: "LEGO® Voyagers",             subtitle: "",                          series: "Upcoming",       tag: "upcoming", appid: 0       },
  { id: 0,       title: "LEGO® Party!",               subtitle: "",                          series: "Upcoming",       tag: "upcoming", appid: 0       },
  { id: 0,       title: "LEGO® Batman:",              subtitle: "Legacy of the Dark Knight", series: "Batman",         tag: "upcoming", appid: 0       },
];

const SERIES_COLORS = {
  'Star Wars': '#ffe81f',
  'Indiana Jones': '#c8873e',
  'Harry Potter': '#740001',
  'Middle-Earth': '#4a7c59',
  'Marvel': '#e62429',
  'DC': '#0476F2',
  'Batman': '#1a1a2e',
  'Jurassic': '#4a7c59',
  'City': '#0066cc',
  'LEGO Movie': '#ff8800',
  'Worlds': '#00c853',
  'Brawls': '#f44336',
  'Builder': '#ff9800',
  'Bricktales': '#9c27b0',
  'Horizon': '#2196f3',
  'Upcoming': '#607d8b',
};

// ─── HERO BRICKS GRID ────────────────────────────────────────
function buildHeroBricksGrid() {
  const grid = document.getElementById('heroBricksGrid');
  const count = Math.ceil(window.innerWidth / 80) * Math.ceil(window.innerHeight / 80);
  for (let i = 0; i < Math.min(count, 200); i++) {
    const b = document.createElement('div');
    b.className = 'hero-brick';
    b.style.cssText = `--i:${i};height:${40 + Math.random()*40}px;animation-duration:${2+Math.random()*3}s;animation-delay:${Math.random()*2}s`;
    grid.appendChild(b);
  }
}

// ─── FLOATING BRICKS ─────────────────────────────────────────
function buildFloatingBricks() {
  const container = document.getElementById('floatBricks');
  const colors = ['#FFD700','#DA291C','#006DB7','#00A650','#FF6B00','#9b59b6'];
  for (let i = 0; i < 18; i++) {
    const b = document.createElement('div');
    b.className = 'fbrick';
    const c = colors[Math.floor(Math.random() * colors.length)];
    b.style.cssText = `
      background:${c};
      left:${Math.random()*100}%;
      width:${30+Math.random()*30}px;
      height:${18+Math.random()*14}px;
      animation-duration:${8+Math.random()*12}s;
      animation-delay:${Math.random()*10}s;
    `;
    container.appendChild(b);
  }
}

// ─── GAME CARDS ───────────────────────────────────────────────
function buildGamesGrid() {
  const grid = document.getElementById('gamesGrid');
  grid.innerHTML = '';
  GAMES.forEach((game, idx) => {
    const card = document.createElement('div');
    card.className = 'game-card';
    card.style.animationDelay = `${idx * 0.05}s`;
    card.onclick = () => openModal(game);

    const imageUrl = game.appid
      ? `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/header.jpg`
      : null;

    const tagClass = game.tag === 'upcoming' ? 'tag-upcoming' : 'tag-action';
    const tagLabel = game.tag === 'upcoming' ? '🔜 À venir' : '🎮 Disponible';

    card.innerHTML = `
      <div class="game-card-image-wrap">
        ${imageUrl
          ? `<img class="game-card-image" src="${imageUrl}" alt="${game.title}" loading="lazy" onerror="this.parentNode.innerHTML=getPlaceholder('${game.series}')">`
          : getPlaceholderHTML(game.series)
        }
        <div class="game-card-overlay"></div>
        <div class="play-btn">
          <svg width="20" height="20" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
        </div>
      </div>
      <div class="game-card-info">
        <div class="game-card-series" style="color:${SERIES_COLORS[game.series] || '#FFD700'}">${game.series}</div>
        <div class="game-card-title">${game.title}${game.subtitle ? '<br><small style="font-size:0.85em;opacity:0.7">'+game.subtitle+'</small>' : ''}</div>
        <div class="game-card-meta">
          <span class="game-tag ${tagClass}">${tagLabel}</span>
          ${game.tag !== 'upcoming' ? `<span class="game-tag tag-steam">Steam</span>` : ''}
        </div>
      </div>
    `;
    grid.appendChild(card);
  });
}

function getPlaceholderHTML(series) {
  const color = SERIES_COLORS[series] || '#333';
  return `<div class="game-card-placeholder" style="background:${color}22;border-bottom:3px solid ${color}44">
    <span style="font-size:2rem">🧱</span>
    <span>${series}</span>
  </div>`;
}

// ─── MODAL ────────────────────────────────────────────────────
const modalOverlay = document.getElementById('modal-overlay');
const modalClose = document.getElementById('modalClose');

modalClose.onclick = closeModal;
modalOverlay.onclick = (e) => { if (e.target === modalOverlay) closeModal(); };
document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeModal(); });

function closeModal() {
  modalOverlay.classList.remove('active');
  document.body.style.overflow = '';
  // stop any iframes
  document.getElementById('modalTrailer').innerHTML = '';
}

async function openModal(game) {
  document.body.style.overflow = 'hidden';
  modalOverlay.classList.add('active');
  modalOverlay.scrollTop = 0;

  // Reset
  document.getElementById('modalSeries').textContent = game.series;
  document.getElementById('modalTitle').textContent = game.title + (game.subtitle ? ': ' + game.subtitle : '');
  document.getElementById('modalDesc').innerHTML = `<div class="loading-spinner"><div class="spinner"></div>Chargement des données Steam...</div>`;
  document.getElementById('modalTrailer').innerHTML = `<div class="no-trailer">Chargement...</div>`;
  document.getElementById('modalScreenshots').innerHTML = '';
  document.getElementById('modalAchievements').innerHTML = `<div class="loading-spinner"><div class="spinner"></div>Chargement des succès...</div>`;

  const heroImg = document.getElementById('modalHero');
  // Remove old img
  const oldImg = heroImg.querySelector('img');
  if (oldImg) oldImg.remove();

  if (game.appid) {
    const steamLink = `https://store.steampowered.com/app/${game.appid}`;
    document.getElementById('modalSteamLink').href = steamLink;
    document.getElementById('modalSteamLink').style.display = 'inline-flex';

    // Hero image
    const img = document.createElement('img');
    img.src = `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/library_hero.jpg`;
    img.onerror = () => { img.src = `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/header.jpg`; };
    img.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;object-fit:cover;';
    heroImg.insertBefore(img, heroImg.firstChild);

    // Fetch game details from Golang backend
    await Promise.all([
      loadGameDetails(game),
      loadAchievements(game),
    ]);
  } else {
    document.getElementById('modalSteamLink').style.display = 'none';
    document.getElementById('modalDesc').innerHTML = `
      <p>Ce jeu n'est pas encore disponible sur Steam. Restez à l'affût des annonces officielles !</p>
      <p style="margin-top:12px;color:rgba(255,165,0,0.8)">🔜 Sortie prévue prochainement.</p>
    `;
    document.getElementById('modalTrailer').innerHTML = `<div class="no-trailer"><span style="font-size:3rem">⏳</span>Trailer bientôt disponible</div>`;
    document.getElementById('modalAchievements').innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px">Succès non disponibles pour les jeux à venir.</p>`;
  }
}

async function loadGameDetails(game) {
  try {
    // Call Golang backend
    const res = await fetch(`${API_BASE_URL}/api/game/${game.appid}`);
    const data = await res.json();

    if (data.success && data.data) {
      const d = data.data;

      // Description
      const desc = d.short_description || d.detailed_description || 'Description non disponible.';
      document.getElementById('modalDesc').innerHTML = desc;

      // Trailer
      const trailerEl = document.getElementById('modalTrailer');
      if (d.movies && d.movies.length > 0) {
        const movie = d.movies[0];
        const mp4 = movie.webm?.max || movie.mp4?.max || '';
        if (mp4) {
          trailerEl.innerHTML = `<video controls autoplay muted loop style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover" src="${mp4}"></video>`;
        } else {
          trailerEl.innerHTML = `<div class="no-trailer"><span style="font-size:3rem">🎬</span>Trailer non disponible</div>`;
        }
      } else {
        const ytQuery = encodeURIComponent(game.title + ' ' + (game.subtitle || '') + ' official trailer');
        trailerEl.innerHTML = `
          <div class="no-trailer" style="flex-direction:column;gap:12px">
            <span style="font-size:3rem">🎬</span>
            <span>Trailer non disponible via Steam</span>
            <a href="https://www.youtube.com/results?search_query=${ytQuery}" target="_blank"
               style="color:var(--lego-yellow);font-size:0.85rem;text-decoration:underline">
              Rechercher sur YouTube ↗
            </a>
          </div>`;
      }

      // Screenshots
      const screensEl = document.getElementById('modalScreenshots');
      if (d.screenshots && d.screenshots.length > 0) {
        screensEl.innerHTML = d.screenshots.slice(0, 12).map(s =>
          `<div class="screenshot" onclick="openScreenshot('${s.path_full}')">
            <img src="${s.path_thumbnail}" alt="screenshot" loading="lazy">
          </div>`
        ).join('');
      } else {
        screensEl.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:16px">Aucun screenshot disponible.</p>`;
      }

    } else {
      throw new Error('No data');
    }
  } catch (e) {
    console.error('Game details error:', e);
    document.getElementById('modalDesc').innerHTML = `
      <p>Les données Steam n'ont pas pu être récupérées.</p>
      <p style="margin-top:8px;font-size:0.85rem;color:rgba(255,255,255,0.4)">
        Assurez-vous que le serveur Golang est démarré (port 8080).
        <a href="https://store.steampowered.com/app/${game.appid}" target="_blank" style="color:var(--lego-yellow)">
          Voir directement sur Steam ↗
        </a>
      </p>
    `;
    document.getElementById('modalTrailer').innerHTML = `<div class="no-trailer"><span style="font-size:3rem">⚠️</span>Données non disponibles</div>`;
    document.getElementById('modalScreenshots').innerHTML = '';
  }
}

async function loadAchievements(game) {
  const el = document.getElementById('modalAchievements');
  try {
    // Call Golang backend for achievements
    const res = await fetch(`${API_BASE_URL}/api/achievements/${game.appid}`);
    const data = await res.json();

    if (!data.success || !data.achievements || data.achievements.length === 0) {
      el.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px 0">Aucun succès disponible pour ce jeu.</p>`;
      return;
    }

    // Sort by percentage desc, show top 18
    const achievements = data.achievements.sort((a, b) => b.percent - a.percent).slice(0, 18);

    el.innerHTML = `<div class="achievements-list">` +
      achievements.map(a => `
        <div class="achievement-item">
          <div class="achievement-icon">
            ${a.icon ? `<img src="${a.icon}" alt="${a.displayName}" onerror="this.parentNode.innerHTML='🏆'">` : '🏆'}
          </div>
          <div class="achievement-info">
            <div class="achievement-name">${a.displayName || a.name}</div>
            <div class="achievement-desc">${a.description || ''}</div>
            <div class="achievement-bar">
              <div class="achievement-bar-fill" style="width:${a.percent.toFixed(1)}%"></div>
            </div>
            <div class="achievement-pct">${a.percent.toFixed(1)}% des joueurs</div>
          </div>
        </div>
      `).join('') +
    `</div>`;

    // Animate bars
    setTimeout(() => {
      document.querySelectorAll('.achievement-bar-fill').forEach(bar => {
        const w = bar.style.width;
        bar.style.width = '0';
        requestAnimationFrame(() => { bar.style.width = w; });
      });
    }, 100);

  } catch (e) {
    console.error('Achievements error:', e);
    el.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px 0">Impossible de charger les succès. Assurez-vous que le serveur Golang est démarré.</p>`;
  }
}

// ─── SCREENSHOT LIGHTBOX ─────────────────────────────────────
function openScreenshot(url) {
  const overlay = document.createElement('div');
  overlay.style.cssText = `position:fixed;inset:0;z-index:9999;background:rgba(0,0,0,0.95);display:flex;align-items:center;justify-content:center;cursor:zoom-out;`;
  const img = document.createElement('img');
  img.src = url;
  img.style.cssText = `max-width:90vw;max-height:90vh;border-radius:8px;box-shadow:0 20px 80px rgba(0,0,0,0.8);`;
  overlay.appendChild(img);
  overlay.onclick = () => document.body.removeChild(overlay);
  document.body.appendChild(overlay);
}

// ─── NOTIFICATIONS ───────────────────────────────────────────
function showNotif(msg) {
  const n = document.getElementById('notif');
  document.getElementById('notifMsg').textContent = msg;
  n.classList.add('show');
  setTimeout(() => n.classList.remove('show'), 3000);
}

// ─── SCROLL ANIMATIONS ───────────────────────────────────────
function initScrollAnimations() {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(e => {
      if (e.isIntersecting) {
        e.target.style.opacity = '1';
        e.target.style.transform = 'translateY(0)';
      }
    });
  }, { threshold: 0.1 });

  document.querySelectorAll('.game-card').forEach((card, i) => {
    card.style.opacity = '0';
    card.style.transition = `opacity 0.5s ${i*0.04}s, transform 0.5s ${i*0.04}s`;
    observer.observe(card);
  });
}

// ─── CURSOR BRICK TRAIL ──────────────────────────────────────
let lastTrail = 0;
document.addEventListener('mousemove', (e) => {
  if (Date.now() - lastTrail < 60) return;
  lastTrail = Date.now();
  const trail = document.createElement('div');
  trail.style.cssText = `
    position:fixed;
    left:${e.clientX - 8}px;
    top:${e.clientY - 5}px;
    width:16px;height:10px;
    border-radius:2px;
    background:${['#FFD700','#DA291C','#006DB7','#00A650'][Math.floor(Math.random()*4)]};
    pointer-events:none;
    z-index:9998;
    opacity:0.8;
    animation:trailFade 0.5s forwards;
  `;
  document.body.appendChild(trail);
  setTimeout(() => trail.remove(), 500);
});

// ─── INIT ────────────────────────────────────────────────────
window.addEventListener('DOMContentLoaded', () => {
  buildHeroBricksGrid();
  buildFloatingBricks();
  buildGamesGrid();
  setTimeout(initScrollAnimations, 100);
});
