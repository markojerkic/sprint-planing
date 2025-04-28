function clickOutsideCallback(event) {
	const isClickOutsideMenu = event.target.closest("[data-open]") === null;
	if (!isClickOutsideMenu) {
		return;
	}

	const mobileMenu = document.querySelector("[data-open]");
	mobileMenu.setAttribute("data-open", "false");
}

function attachMobileMenu() {
	const mobileMenu = document.querySelector("[data-open]");
	const mobileMenuButton = mobileMenu.querySelector("button");

	mobileMenuButton.addEventListener("click", () => {
		const currentState = mobileMenu.getAttribute("data-open");
		mobileMenu.setAttribute(
			"data-open",
			currentState === "true" ? "false" : "true",
		);
	});

	document.addEventListener("click", clickOutsideCallback);
}

document.addEventListener("DOMContentLoaded", attachMobileMenu);
// Htmx after boosted swap
document.addEventListener("htmx:afterSwap", (e) => {
	const isBoosted = e.detail.requestConfig?.boosted;
	if (isBoosted) {
		attachMobileMenu();
	}
});
