var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

document.addEventListener('DOMContentLoaded', function () {
  function fadeOutElement() {
      console.log('Fading out flash element');
      const flashElement = document.getElementById('flash');
      if (flashElement) {
          setTimeout(() => {
              flashElement.style.opacity = 0;
              setTimeout(() => {
                  flashElement.remove();
              }, 500);
          }, 5000);
      }
  }
  fadeOutElement();
});
