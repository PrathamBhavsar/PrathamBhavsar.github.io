const root = document.documentElement;
const themeToggle = document.getElementById("themeToggle");
const menuToggle = document.getElementById("menuToggle");
const mobileMenu = document.getElementById("mobileMenu");

lucide.createIcons();

const year = document.getElementById("currentYear");
if (year) year.textContent = new Date().getFullYear();

/* ---------- theme ---------- */

const storedTheme = localStorage.getItem("theme");
if (storedTheme) root.setAttribute("data-theme", storedTheme);

const updateToggleIcon = () => {
  const isDark = root.getAttribute("data-theme") !== "light";
  themeToggle.querySelector(".icon").textContent = isDark ? "◐" : "☀";
};
updateToggleIcon();

themeToggle.addEventListener("click", () => {
  const next = root.getAttribute("data-theme") === "light" ? "dark" : "light";
  root.setAttribute("data-theme", next);
  localStorage.setItem("theme", next);
  updateToggleIcon();
});

/* ---------- mobile nav ---------- */

if (menuToggle) {
  menuToggle.addEventListener("click", () => {
    const open = mobileMenu.classList.toggle("open");
    menuToggle.setAttribute("aria-expanded", String(open));
  });
  mobileMenu.querySelectorAll("a").forEach((a) =>
    a.addEventListener("click", () => {
      mobileMenu.classList.remove("open");
      menuToggle.setAttribute("aria-expanded", "false");
    })
  );
}

/* ---------- shared helpers ---------- */

const esc = (s) =>
  s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

const inline = (s) =>
  esc(s)
    .replace(/`([^`]+)`/g, "<code>$1</code>")
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");

const linkIcon = { live: "globe", source: "git-branch", demo: "play" };

// Media well: real GIF when captured, an honest placeholder when not.
function mediaWell(p) {
  const { demo, poster, aspect } = p.media;
  if (!demo) {
    return `<div class="well is-empty" data-aspect="${aspect}">
      <span class="well-empty-label">demo · pending capture</span>
      <span class="well-empty-meta">${aspect === "phone" ? "mobile" : "desktop"} · gif</span>
    </div>`;
  }
  return `<div class="well" data-aspect="${aspect}">
    <img class="well-img" src="${poster || demo}" data-still="${poster || demo}"
         data-motion="${demo}" alt="${esc(p.title)} demo" loading="lazy" decoding="async" />
  </div>`;
}

function linkRow(p) {
  const badges = p.links
    .map(
      (l) => `<a class="badge badge-${l.kind}" href="${l.href}" target="_blank" rel="noopener noreferrer">
        <i data-lucide="${linkIcon[l.kind]}"></i>${esc(l.label)}</a>`
    )
    .join("");
  const closed = p.links.length
    ? ""
    : `<span class="badge badge-closed"><i data-lucide="lock"></i>private · demo only</span>`;
  return badges + closed;
}

/* ---------- project index (home) ---------- */

const indexEl = document.getElementById("projectIndex");

if (indexEl) {
  fetch("projects.json")
    .then((r) => r.json())
    .then((projects) => {
      indexEl.innerHTML = projects
        .map(
          (p) => `
        <article class="entry">
          <a class="entry-media" href="project.html?p=${p.slug}" aria-label="${esc(p.title)} case study">
            ${mediaWell(p)}
          </a>
          <div class="entry-body">
            <p class="entry-id">${p.slug} <span class="entry-role">${p.role}</span></p>
            <h3 class="entry-title"><a href="project.html?p=${p.slug}">${esc(p.title)}</a></h3>
            <p class="entry-line">${esc(p.one_liner)}</p>
            <p class="entry-note">${esc(p.note)}</p>
            <p class="spec">${p.spec.map(esc).join(" <span>·</span> ")}</p>
            <p class="stack-row">${p.stack.map((s) => `<span class="chip">${esc(s)}</span>`).join("")}</p>
            <p class="entry-links">
              ${linkRow(p)}
              <a class="badge badge-read" href="project.html?p=${p.slug}">Read case study <i data-lucide="arrow-right"></i></a>
            </p>
          </div>
        </article>`
        )
        .join("");
      lucide.createIcons();
      wakeOnView(indexEl);
    })
    .catch(() => {
      indexEl.innerHTML = `<p class="load-error">Couldn't load projects. <a href="https://github.com/PrathamBhavsar">Browse GitHub instead</a>.</p>`;
    });
}

// The one authored moment: an entry wakes when you reach it, and its demo
// only animates while it is on screen.
function wakeOnView(scope) {
  const entries = [...scope.querySelectorAll(".entry")];
  if (!("IntersectionObserver" in window) || matchMedia("(prefers-reduced-motion: reduce)").matches) {
    entries.forEach((e) => e.classList.add("awake"));
    return;
  }
  const io = new IntersectionObserver(
    (records) =>
      records.forEach((r) => {
        if (r.isIntersecting) r.target.classList.add("awake");
        const img = r.target.querySelector(".well-img");
        if (img) img.src = r.isIntersecting ? img.dataset.motion : img.dataset.still;
      }),
    { rootMargin: "-12% 0px -12% 0px" }
  );
  entries.forEach((e) => io.observe(e));
}

/* ---------- case study page ---------- */

const studyEl = document.getElementById("study");

if (studyEl) {
  const slug = new URLSearchParams(location.search).get("p") || "";
  const safe = /^[a-z0-9-]+$/.test(slug);

  Promise.all([
    fetch("projects.json").then((r) => r.json()),
    safe ? fetch(`content/${slug}.md`).then((r) => (r.ok ? r.text() : Promise.reject())) : Promise.reject(),
  ])
    .then(([projects, md]) => {
      const p = projects.find((x) => x.slug === slug);
      if (!p) throw new Error("unknown project");
      document.title = `${p.title} | Pratham Bhavsar`;
      studyEl.innerHTML = `
        <p class="entry-id">${p.slug} <span class="entry-role">${p.role}</span></p>
        <h1 class="study-title">${esc(p.title)}</h1>
        <p class="study-line">${esc(p.one_liner)}</p>
        <div class="study-cols">
          <aside class="study-rail">
            ${mediaWell(p)}
            <dl class="rail-meta">
              <dt>Scope</dt><dd>${p.spec.map(esc).join("<br>")}</dd>
              <dt>Stack</dt><dd>${p.stack.map(esc).join("<br>")}</dd>
            </dl>
            <p class="entry-links">${linkRow(p)}</p>
          </aside>
          <div class="prose">${renderMd(md)}</div>
        </div>`;
      lucide.createIcons();
    })
    .catch(() => {
      studyEl.innerHTML = `<h1 class="study-title">Not found</h1>
        <p class="study-line">No case study for <code>${esc(slug)}</code>.</p>
        <p class="entry-links"><a class="badge badge-read" href="index.html#projects">All projects <i data-lucide="arrow-right"></i></a></p>`;
      lucide.createIcons();
    });
}

// Minimal renderer for the subset these case studies use:
// ## heading, paragraph, - bullet, 1. item, **bold**, `code`.
function renderMd(md) {
  return md
    .trim()
    .split(/\n{2,}/)
    .map((block) => {
      const lines = block.split("\n");
      if (/^## /.test(block)) return `<h2>${inline(block.slice(3))}</h2>`;
      if (lines.every((l) => /^- /.test(l)))
        return `<ul>${lines.map((l) => `<li>${inline(l.slice(2))}</li>`).join("")}</ul>`;
      if (lines.every((l) => /^\d+\. /.test(l)))
        return `<ol>${lines.map((l) => `<li>${inline(l.replace(/^\d+\.\s/, ""))}</li>`).join("")}</ol>`;
      return `<p>${inline(block.replace(/\n/g, " "))}</p>`;
    })
    .join("");
}
