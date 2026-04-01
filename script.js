// Client script for LEGO Games Wiki
// - Initializes Supabase client after DOMContentLoaded
// - Provides signUp/login/logout/getUser functions
// - Builds games grid and modal behavior
// - Uses API_BASE_URL = window.location.origin for backend calls

(() => {
  const API_BASE_URL = window.location.origin;
  let supabaseClient = null;

  // ----- Utilities -----
  function el(selector) {
    return document.querySelector(selector);
  }
  function byId(id) {
    return document.getElementById(id);
  }
  function showNotif(msg) {
    const n = byId("notif");
    if (!n) return;
    byId("notifMsg").textContent = msg;
    n.classList.add("show");
    setTimeout(() => n.classList.remove("show"), 3000);
  }

  // ----- Auth helpers (will be bound to window) -----
  async function signUp(email, password) {
    if (!supabaseClient) {
      console.error("Supabase client not initialized");
      showNotif("Erreur: auth non initialisée");
      return;
    }
    const { data, error } = await supabaseClient.auth.signUp({
      email,
      password,
    });
    if (error) {
      console.error(error);
      showNotif("Inscription impossible: " + (error.message || error));
      return null;
    }
    showNotif("Compte créé — vérifie ta boîte mail");
    return data;
  }

  async function login(email, password) {
    if (!supabaseClient) {
      console.error("Supabase client not initialized");
      showNotif("Erreur: auth non initialisée");
      return;
    }
    const { data, error } = await supabaseClient.auth.signInWithPassword({
      email,
      password,
    });
    if (error) {
      console.error(error);
      showNotif("Connexion impossible: " + (error.message || error));
      return null;
    }
    showNotif("Connecté");
    updateAuthUI();
    return data;
  }

  async function logout() {
    if (!supabaseClient) return;
    await supabaseClient.auth.signOut();
    showNotif("Déconnecté");
    updateAuthUI();
  }

  async function getUser() {
    if (!supabaseClient) return null;
    const { data } = await supabaseClient.auth.getUser();
    return data?.user || null;
  }

  // Called by the login button in index.html
  async function loginUser() {
    const email = byId("email")?.value || "";
    const password = byId("password")?.value || "";
    if (!email || !password) {
      showNotif("Email et mot de passe requis");
      return;
    }
    await login(email, password);
  }

  // Expose functions globally (index.html calls loginUser())
  window.signUp = signUp;
  window.login = login;
  window.logout = logout;
  window.getUser = getUser;
  window.loginUser = loginUser;

  // ----- Supabase & Auth UI management -----
  function createAuthStatusElement() {
    // Append a small auth status area to nav if not present
    let status = byId("authStatus");
    if (status) return status;
    const navLinks = document.querySelector(".nav-links");
    status = document.createElement("div");
    status.id = "authStatus";
    status.style.marginLeft = "12px";
    status.style.color = "white";
    status.style.fontSize = "0.9rem";
    status.style.display = "flex";
    status.style.gap = "8px";
    status.style.alignItems = "center";

    const btn = document.createElement("button");
    btn.id = "authBtn";
    btn.style.cursor = "pointer";
    btn.style.padding = "6px 10px";
    btn.style.borderRadius = "6px";
    btn.style.border = "none";
    btn.style.background = "var(--lego-yellow)";
    btn.style.color = "#111";
    btn.style.fontWeight = "700";
    btn.textContent = "Connexion";

    btn.onclick = async () => {
      const user = await getUser();
      if (user) {
        // logout
        await logout();
      } else {
        // focus email input to help quick login
        const emailInput = byId("email");
        if (emailInput) emailInput.focus();
        showNotif("Saisis ton e‑mail et mot de passe puis clique sur Login");
      }
    };

    status.appendChild(btn);
    if (navLinks) navLinks.appendChild(status);
    return status;
  }

  async function updateAuthUI() {
    const status = createAuthStatusElement();
    const user = await getUser();
    const btn = byId("authBtn");
    if (user) {
      status.title = user.email;
      btn.textContent = "Déconnexion";
      // show a short welcome
      const welcome = status.querySelector(".welcome");
      if (!welcome) {
        const w = document.createElement("span");
        w.className = "welcome";
        w.style.opacity = "0.9";
        w.style.fontSize = "0.85rem";
        w.textContent = user.email.split("@")[0];
        status.insertBefore(w, btn);
      } else {
        welcome.textContent = user.email.split("@")[0];
      }
    } else {
      status.title = "Non connecté";
      if (btn) btn.textContent = "Connexion";
      const welcome = status.querySelector(".welcome");
      if (welcome) welcome.remove();
    }
  }

  // ----- Games and UI (kept mostly compatible with original) -----
  const GAMES = [
    {
      id: 32440,
      title: "LEGO® Star Wars™",
      subtitle: "The Complete Saga",
      series: "Star Wars",
      tag: "action",
      appid: 32440,
    },
    {
      id: 32510,
      title: "LEGO® Star Wars™ III",
      subtitle: "The Clone Wars",
      series: "Star Wars",
      tag: "action",
      appid: 32510,
    },
    {
      id: 438640,
      title: "LEGO® Star Wars™",
      subtitle: "The Force Awakens",
      series: "Star Wars",
      tag: "action",
      appid: 438640,
    },
    {
      id: 920210,
      title: "LEGO® Star Wars™",
      subtitle: "The Skywalker Saga",
      series: "Star Wars",
      tag: "action",
      appid: 920210,
    },
    {
      id: 32330,
      title: "LEGO® Indiana Jones™",
      subtitle: "The Original Adventures",
      series: "Indiana Jones",
      tag: "aventure",
      appid: 32330,
    },
    {
      id: 32450,
      title: "LEGO® Indiana Jones™ 2",
      subtitle: "The Adventure Continues",
      series: "Indiana Jones",
      tag: "aventure",
      appid: 32450,
    },
    {
      id: 21130,
      title: "LEGO® Harry Potter™",
      subtitle: "Years 1-4",
      series: "Harry Potter",
      tag: "magie",
      appid: 21130,
    },
    {
      id: 204120,
      title: "LEGO® Harry Potter™",
      subtitle: "Years 5-7",
      series: "Harry Potter",
      tag: "magie",
      appid: 204120,
    },
    {
      id: 214510,
      title: "LEGO® The Lord of the Rings™",
      subtitle: "",
      series: "Middle-Earth",
      tag: "aventure",
      appid: 214510,
    },
    {
      id: 292712,
      title: "LEGO® The Hobbit™",
      subtitle: "",
      series: "Middle-Earth",
      tag: "aventure",
      appid: 292712,
    },
    {
      id: 249130,
      title: "LEGO® Marvel™ Super Heroes",
      subtitle: "",
      series: "Marvel",
      tag: "action",
      appid: 249130,
    },
    {
      id: 647830,
      title: "LEGO® Marvel™ Super Heroes 2",
      subtitle: "",
      series: "Marvel",
      tag: "action",
      appid: 647830,
    },
    {
      id: 405310,
      title: "LEGO® Marvel's Avengers",
      subtitle: "",
      series: "Marvel",
      tag: "action",
      appid: 405310,
    },
    {
      id: 829110,
      title: "LEGO® DC Super-Villains",
      subtitle: "",
      series: "DC",
      tag: "action",
      appid: 829110,
    },
    {
      id: 21000,
      title: "LEGO® Batman™",
      subtitle: "The Videogame",
      series: "Batman",
      tag: "action",
      appid: 21000,
    },
    {
      id: 213330,
      title: "LEGO® Batman™ 2",
      subtitle: "DC Super Heroes",
      series: "Batman",
      tag: "action",
      appid: 213330,
    },
    {
      id: 313690,
      title: "LEGO® Batman™ 3",
      subtitle: "Beyond Gotham",
      series: "Batman",
      tag: "action",
      appid: 313690,
    },
    {
      id: 352400,
      title: "LEGO® Jurassic World",
      subtitle: "",
      series: "Jurassic",
      tag: "aventure",
      appid: 352400,
    },
    {
      id: 578330,
      title: "LEGO® City Undercover",
      subtitle: "",
      series: "City",
      tag: "sandbox",
      appid: 578330,
    },
    {
      id: 267530,
      title: "The LEGO® Movie",
      subtitle: "Videogame",
      series: "LEGO Movie",
      tag: "aventure",
      appid: 267530,
    },
    {
      id: 881320,
      title: "The LEGO® Movie 2",
      subtitle: "Videogame",
      series: "LEGO Movie",
      tag: "aventure",
      appid: 881320,
    },
    {
      id: 332310,
      title: "LEGO® Worlds",
      subtitle: "",
      series: "Worlds",
      tag: "sandbox",
      appid: 332310,
    },
    {
      id: 1731460,
      title: "LEGO® Brawls",
      subtitle: "",
      series: "Brawls",
      tag: "combat",
      appid: 1731460,
    },
    {
      id: 1544360,
      title: "LEGO® Builder's Journey",
      subtitle: "",
      series: "Builder",
      tag: "puzzle",
      appid: 1544360,
    },
    {
      id: 1898290,
      title: "LEGO® Bricktales",
      subtitle: "",
      series: "Bricktales",
      tag: "puzzle",
      appid: 1898290,
    },
    {
      id: 2428810,
      title: "LEGO® Horizon Adventures",
      subtitle: "",
      series: "Horizon",
      tag: "action",
      appid: 2428810,
    },
    {
      id: 1538550,
      title: "LEGO® Voyagers",
      subtitle: "",
      series: "Bricktales",
      tag: "puzzle",
      appid: 0,
    },
    {
      id: 1969370,
      title: "LEGO® Party!",
      subtitle: "",
      series: "Brawls",
      tag: "puzzle",
      appid: 0,
    },
    {
      id: 2215200,
      title: "LEGO® Batman:",
      subtitle: "Legacy of the Dark Knight",
      series: "Batman",
      tag: "upcoming",
      appid: 0,
    },
  ];

  const SERIES_COLORS = {
    "Star Wars": "#ffe81f",
    "Indiana Jones": "#c8873e",
    "Harry Potter": "#740001",
    "Middle-Earth": "#4a7c59",
    Marvel: "#e62429",
    DC: "#0476F2",
    Batman: "#1a1a2e",
    Jurassic: "#4a7c59",
    City: "#0066cc",
    "LEGO Movie": "#ff8800",
    Worlds: "#00c853",
    Brawls: "#f44336",
    Builder: "#ff9800",
    Bricktales: "#9c27b0",
    Horizon: "#2196f3",
    Upcoming: "#607d8b",
  };

  // Build hero bricks and floating bricks
  function buildHeroBricksGrid() {
    const grid = byId("heroBricksGrid");
    if (!grid) return;
    const count =
      Math.ceil(window.innerWidth / 80) * Math.ceil(window.innerHeight / 80);
    for (let i = 0; i < Math.min(count, 200); i++) {
      const b = document.createElement("div");
      b.className = "hero-brick";
      b.style.cssText = `--i:${i};height:${40 + Math.random() * 40}px;animation-duration:${2 + Math.random() * 3}s;animation-delay:${Math.random() * 2}s`;
      grid.appendChild(b);
    }
  }

  function buildFloatingBricks() {
    const container = byId("floatBricks");
    if (!container) return;
    const colors = [
      "#FFD700",
      "#DA291C",
      "#006DB7",
      "#00A650",
      "#FF6B00",
      "#9b59b6",
    ];
    for (let i = 0; i < 18; i++) {
      const b = document.createElement("div");
      b.className = "fbrick";
      const c = colors[Math.floor(Math.random() * colors.length)];
      b.style.cssText = `
        background:${c};
        left:${Math.random() * 100}%;
        width:${30 + Math.random() * 30}px;
        height:${18 + Math.random() * 14}px;
        animation-duration:${8 + Math.random() * 12}s;
        animation-delay:${Math.random() * 10}s;
      `;
      container.appendChild(b);
    }
  }

  function getPlaceholderHTML(series) {
    const color = SERIES_COLORS[series] || "#333";
    return `<div class="game-card-placeholder" style="background:${color}22;border-bottom:3px solid ${color}44">
      <span style="font-size:2rem">🧱</span>
      <span>${series}</span>
    </div>`;
  }

  // Build games grid
  function buildGamesGrid() {
    const grid = byId("gamesGrid");
    if (!grid) return;
    grid.innerHTML = "";
    GAMES.forEach((game, idx) => {
      const card = document.createElement("div");
      card.className = "game-card";
      card.style.animationDelay = `${idx * 0.05}s`;
      card.onclick = () => openModal(game);

      const imageUrl = game.appid
        ? `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/header.jpg`
        : null;

      const tagClass = game.tag === "upcoming" ? "tag-upcoming" : "tag-action";
      const tagLabel = game.tag === "upcoming" ? "🔜 À venir" : "🎮 Disponible";

      card.innerHTML = `
        <div class="game-card-image-wrap">
          ${
            imageUrl
              ? `<img class="game-card-image" src="${imageUrl}" alt="${game.title}" loading="lazy" onerror="this.parentNode.innerHTML=getPlaceholderHTML('${game.series}')">`
              : getPlaceholderHTML(game.series)
          }
          <div class="game-card-overlay"></div>
          <div class="play-btn">
            <svg width="20" height="20" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
          </div>
        </div>
        <div class="game-card-info">
          <div class="game-card-series" style="color:${SERIES_COLORS[game.series] || "#FFD700"}">${game.series}</div>
          <div class="game-card-title">${game.title}${game.subtitle ? '<br><small style="font-size:0.85em;opacity:0.7">' + game.subtitle + "</small>" : ""}</div>
          <div class="game-card-meta">
            <span class="game-tag ${tagClass}">${tagLabel}</span>
            ${game.tag !== "upcoming" ? `<span class="game-tag tag-steam">Steam</span>` : ""}
          </div>
        </div>
      `;
      grid.appendChild(card);
    });
  }

  // Modal controls (reused references)
  const modalOverlay = byId("modal-overlay");
  const modalClose = byId("modalClose");
  if (modalClose) modalClose.onclick = closeModal;
  if (modalOverlay)
    modalOverlay.onclick = (e) => {
      if (e.target === modalOverlay) closeModal();
    };
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") closeModal();
  });

  function closeModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.remove("active");
    document.body.style.overflow = "";
    const mt = byId("modalTrailer");
    if (mt) mt.innerHTML = "";
  }

  async function openModal(game) {
    document.body.style.overflow = "hidden";
    if (modalOverlay) modalOverlay.classList.add("active");
    if (modalOverlay) modalOverlay.scrollTop = 0;

    const seriesEl = byId("modalSeries");
    const titleEl = byId("modalTitle");
    const descEl = byId("modalDesc");
    const trailerEl = byId("modalTrailer");
    const screensEl = byId("modalScreenshots");
    const achEl = byId("modalAchievements");
    const steamLinkEl = byId("modalSteamLink");

    if (seriesEl) seriesEl.textContent = game.series;
    if (titleEl)
      titleEl.textContent =
        game.title + (game.subtitle ? ": " + game.subtitle : "");
    if (descEl)
      descEl.innerHTML = `<div class="loading-spinner"><div class="spinner"></div>Chargement des données Steam...</div>`;
    if (trailerEl)
      trailerEl.innerHTML = `<div class="no-trailer">Chargement...</div>`;
    if (screensEl) screensEl.innerHTML = "";
    if (achEl)
      achEl.innerHTML = `<div class="loading-spinner"><div class="spinner"></div>Chargement des succès...</div>`;

    const heroImg = byId("modalHero");
    if (heroImg) {
      const oldImg = heroImg.querySelector("img");
      if (oldImg) oldImg.remove();
    }

    if (game.appid) {
      const steamLink = `https://store.steampowered.com/app/${game.appid}`;
      if (steamLinkEl) {
        steamLinkEl.href = steamLink;
        steamLinkEl.style.display = "inline-flex";
      }

      if (heroImg) {
        const img = document.createElement("img");
        img.src = `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/library_hero.jpg`;
        img.onerror = () => {
          img.src = `https://cdn.akamai.steamstatic.com/steam/apps/${game.appid}/header.jpg`;
        };
        img.style.cssText =
          "position:absolute;inset:0;width:100%;height:100%;object-fit:cover;";
        heroImg.insertBefore(img, heroImg.firstChild);
      }

      // Load details and achievements in parallel
      await Promise.all([loadGameDetails(game), loadAchievements(game)]);
    } else {
      if (steamLinkEl) steamLinkEl.style.display = "none";
      if (descEl)
        descEl.innerHTML = `
        <p>Ce jeu n'est pas encore disponible sur Steam. Restez à l'affût des annonces officielles !</p>
        <p style="margin-top:12px;color:rgba(255,165,0,0.8)">🔜 Sortie prévue prochainement.</p>
      `;
      if (trailerEl)
        trailerEl.innerHTML = `<div class="no-trailer"><span style="font-size:3rem">⏳</span>Trailer bientôt disponible</div>`;
      if (achEl)
        achEl.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px">Succès non disponibles pour les jeux à venir.</p>`;
    }
  }

  async function loadGameDetails(game) {
    try {
      const res = await fetch(`${API_BASE_URL}/api/game/${game.appid}`);
      if (!res.ok) throw new Error("Steam API proxy error: " + res.status);
      const data = await res.json();
      if (!data || !data.data) throw new Error("No game data");

      const d = data.data;
      const desc =
        d.short_description ||
        d.detailed_description ||
        "Description non disponible.";
      const descEl = byId("modalDesc");
      if (descEl) descEl.innerHTML = desc;

      // Trailer handling
      const trailerEl = byId("modalTrailer");
      if (trailerEl) {
        if (d.movies && d.movies.length > 0) {
          const movie = d.movies[0];
          const mp4 =
            (movie.webm && movie.webm.max) ||
            (movie.mp4 && movie.mp4.max) ||
            "";
          if (mp4) {
            trailerEl.innerHTML = `<video controls autoplay muted loop style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover" src="${mp4}"></video>`;
          } else {
            trailerEl.innerHTML = `<div class="no-trailer"><span style="font-size:3rem">🎬</span>Trailer non disponible</div>`;
          }
        } else {
          const ytQuery = encodeURIComponent(
            game.title + " " + (game.subtitle || "") + " official trailer",
          );
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
      }

      // Screenshots
      const screensEl = byId("modalScreenshots");
      if (screensEl) {
        if (d.screenshots && d.screenshots.length > 0) {
          screensEl.innerHTML = d.screenshots
            .slice(0, 12)
            .map(
              (s) =>
                `<div class="screenshot" onclick="openScreenshot('${s.path_full}')">
              <img src="${s.path_thumbnail}" alt="screenshot" loading="lazy">
            </div>`,
            )
            .join("");
        } else {
          screensEl.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:16px">Aucun screenshot disponible.</p>`;
        }
      }
    } catch (e) {
      console.error("Game details error:", e);
      const descEl = byId("modalDesc");
      if (descEl)
        descEl.innerHTML = `
        <p>Les données Steam n'ont pas pu être récupérées.</p>
        <p style="margin-top:8px;font-size:0.85rem;color:rgba(255,255,255,0.4)">
          Assurez-vous que le serveur Golang est démarré (port 8080).
          <a href="https://store.steampowered.com/app/${game.appid}" target="_blank" style="color:var(--lego-yellow)">
            Voir directement sur Steam ↗
          </a>
        </p>
      `;
      const trailerEl = byId("modalTrailer");
      if (trailerEl)
        trailerEl.innerHTML = `<div class="no-trailer"><span style="font-size:3rem">⚠️</span>Données non disponibles</div>`;
      const screensEl = byId("modalScreenshots");
      if (screensEl) screensEl.innerHTML = "";
    }
  }

  async function loadAchievements(game) {
    const el = byId("modalAchievements");
    if (!el) return;
    try {
      const res = await fetch(`${API_BASE_URL}/api/achievements/${game.appid}`);
      if (!res.ok) {
        el.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px 0">Aucun succès disponible pour ce jeu.</p>`;
        return;
      }
      const data = await res.json();
      if (
        !data.success ||
        !data.achievements ||
        data.achievements.length === 0
      ) {
        el.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px 0">Aucun succès disponible pour ce jeu.</p>`;
        return;
      }

      const achievements = data.achievements
        .sort((a, b) => b.percent - a.percent)
        .slice(0, 18);

      el.innerHTML =
        `<div class="achievements-list">` +
        achievements
          .map(
            (a) => `
          <div class="achievement-item">
            <div class="achievement-icon">
              ${a.icon ? `<img src="${a.icon}" alt="${a.displayName}" onerror="this.parentNode.innerHTML='🏆'">` : "🏆"}
            </div>
            <div class="achievement-info">
              <div class="achievement-name">${a.displayName || a.name}</div>
              <div class="achievement-desc">${a.description || ""}</div>
              <div class="achievement-bar">
                <div class="achievement-bar-fill" style="width:${(a.percent || 0).toFixed(1)}%"></div>
              </div>
              <div class="achievement-pct">${(a.percent || 0).toFixed(1)}% des joueurs</div>
            </div>
          </div>
        `,
          )
          .join("") +
        `</div>`;

      // Animate bars
      setTimeout(() => {
        document.querySelectorAll(".achievement-bar-fill").forEach((bar) => {
          const w = bar.style.width;
          bar.style.width = "0";
          requestAnimationFrame(() => {
            bar.style.width = w;
          });
        });
      }, 100);
    } catch (e) {
      console.error("Achievements error:", e);
      el.innerHTML = `<p style="color:rgba(255,255,255,0.4);padding:20px 0">Impossible de charger les succès. Assurez-vous que le serveur Golang est démarré.</p>`;
    }
  }

  // Lightbox screenshot opener (exposed globally as the markup uses onclick)
  window.openScreenshot = function (url) {
    const overlay = document.createElement("div");
    overlay.style.cssText = `position:fixed;inset:0;z-index:9999;background:rgba(0,0,0,0.95);display:flex;align-items:center;justify-content:center;cursor:zoom-out;`;
    const img = document.createElement("img");
    img.src = url;
    img.style.cssText = `max-width:90vw;max-height:90vh;border-radius:8px;box-shadow:0 20px 80px rgba(0,0,0,0.8);`;
    overlay.appendChild(img);
    overlay.onclick = () => document.body.removeChild(overlay);
    document.body.appendChild(overlay);
  };

  // Scroll animations
  function initScrollAnimations() {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((e) => {
          if (e.isIntersecting) {
            e.target.style.opacity = "1";
            e.target.style.transform = "translateY(0)";
          }
        });
      },
      { threshold: 0.1 },
    );

    document.querySelectorAll(".game-card").forEach((card, i) => {
      card.style.opacity = "0";
      card.style.transition = `opacity 0.5s ${i * 0.04}s, transform 0.5s ${i * 0.04}s`;
      observer.observe(card);
    });
  }

  // Cursor trail
  let lastTrail = 0;
  document.addEventListener("mousemove", (e) => {
    if (Date.now() - lastTrail < 60) return;
    lastTrail = Date.now();
    const trail = document.createElement("div");
    trail.style.cssText = `
      position:fixed;
      left:${e.clientX - 8}px;
      top:${e.clientY - 5}px;
      width:16px;height:10px;
      border-radius:2px;
      background:${["#FFD700", "#DA291C", "#006DB7", "#00A650"][Math.floor(Math.random() * 4)]};
      pointer-events:none;
      z-index:9998;
      opacity:0.8;
      animation:trailFade 0.5s forwards;
    `;
    document.body.appendChild(trail);
    setTimeout(() => trail.remove(), 500);
  });

  // Init after DOM ready
  window.addEventListener("DOMContentLoaded", async () => {
    // Create supabase client here (only if global `supabase` library is available)
    try {
      if (typeof supabase !== "undefined" && supabase.createClient) {
        // These values come from index.html's included config in the original project.
        // If you want to use your own Supabase instance, replace the URL and key below or
        // set them via environment/config before starting.
        const SUPABASE_URL = "https://sllxqbnjofnzabeidphe.supabase.co";
        const SUPABASE_KEY = "sb_publishable_qvGzCLc9i0gLKEPnGk27Bg_GM99NVOh";
        supabaseClient = supabase.createClient(SUPABASE_URL, SUPABASE_KEY);
      } else {
        console.warn("Supabase library not found; auth will be disabled.");
      }
    } catch (e) {
      console.error("Failed to initialize supabase client:", e);
    }

    // Build UI
    buildHeroBricksGrid();
    buildFloatingBricks();
    buildGamesGrid();

    // small delay then init scroll animations (gives DOM time to insert cards)
    setTimeout(initScrollAnimations, 100);

    // initialize auth UI
    createAuthStatusElement();
    updateAuthUI();

    // Listen to auth changes (if supabase client available)
    if (
      supabaseClient &&
      supabaseClient.auth &&
      supabaseClient.auth.onAuthStateChange
    ) {
      supabaseClient.auth.onAuthStateChange((event, session) => {
        console.log("Auth event:", event);
        updateAuthUI();
      });
    }
  });
})();
