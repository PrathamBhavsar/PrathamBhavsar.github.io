const themeToggle = document.getElementById("themeToggle");
const menuToggle = document.getElementById("menuToggle");
const mobileMenu = document.getElementById("mobileMenu");
const root = document.documentElement;

// Initialize Lucide icons
lucide.createIcons();

// Set current year in footer
document.getElementById("currentYear").textContent = new Date().getFullYear();

const storedTheme = localStorage.getItem("theme");
if (storedTheme) {
  root.setAttribute("data-theme", storedTheme);
}

const updateToggleIcon = () => {
  const isDark = root.getAttribute("data-theme") !== "light";
  themeToggle.querySelector(".icon").textContent = isDark ? "◐" : "☀";
};

updateToggleIcon();

themeToggle.addEventListener("click", () => {
  const current = root.getAttribute("data-theme");
  const next = current === "light" ? "dark" : "light";
  root.setAttribute("data-theme", next);
  localStorage.setItem("theme", next);
  updateToggleIcon();
});

menuToggle.addEventListener("click", () => {
  const isOpen = mobileMenu.style.display === "flex";
  mobileMenu.style.display = isOpen ? "none" : "flex";
});

mobileMenu.querySelectorAll("a").forEach((link) => {
  link.addEventListener("click", () => {
    mobileMenu.style.display = "none";
  });
});