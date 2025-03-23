/** @param {string} str */
function isJson(str) {
	try {
		JSON.parse(str);
	} catch (e) {
		return false;
	}
	return true;
}

document.addEventListener(
	"htmx:wsBeforeMessage",
	/** @param {CustomEvent} e */
	function (e) {
		/** @type {string} */
		const message = e.detail.message;
		if (!isJson(message)) {
			return;
		}
		const json = JSON.parse(message);
		const ticketId = json["ticketID"];
		const isHidden = json["isHidden"];

		/** @type {HTMLElement} */
		const ticketElement = document.querySelector(
			`div[data-ticket-id="${ticketId}"]`,
		);
		const isOwner = ticketElement.getAttribute("data-is-owner") === "true";

		if (isHidden && !isOwner) {
			ticketElement.style.display = "none";
		} else if (!isOwner) {
			// Move to top of the list
			// display flex
			// add animation and scroll into view
			const parentElement = ticketElement.parentElement;
			parentElement.removeChild(ticketElement);
			parentElement.prepend(ticketElement);
			ticketElement.style.display = "flex";
			ticketElement.classList.add("highlight-animation");
			ticketElement.scrollIntoView({ behavior: "smooth" });
			setTimeout(() => {
				ticketElement.classList.remove("highlight-animation");
			}, 1500);
		}
	},
);

// After new ticket swapped, scroll to the new ticket and flash it
document.addEventListener(
	"htmx:oobAfterSwap",
	/** @param {CustomEvent} event */
	function (event) {
		/* @type {HTMLElement} */
		const target = event.detail.target.children[0];

		// Only if target contains attribute data-ticket-id
		if (!target?.hasAttribute("data-ticket-id")) {
			return;
		}

		// Scroll to the new ticket
		target.scrollIntoView({ behavior: "smooth" });

		// Save original border if any
		const originalBorder = target.style.border;

		// Add a CSS class for animation instead of inline styles
		target.classList.add("highlight-animation");

		// Remove the class after animation completes
		setTimeout(() => {
			target.classList.remove("highlight-animation");
			if (originalBorder) {
				target.style.border = originalBorder;
			}
		}, 1500);
	},
);

// Format created on date
function formatDate() {
	const dateElement = document.querySelector("time");
	if (dateElement) {
		const date = new Date(dateElement.getAttribute("datetime"));
		dateElement.textContent = date.toLocaleString();
	}
}
// After boosted link change
document.addEventListener("htmx:afterSwap", formatDate);
// On page load
formatDate();
