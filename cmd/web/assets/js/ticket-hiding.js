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
		console.log("wsBeforeMessage", e);
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
		console.log(ticketElement);
		console.log("isOwner", isOwner);

		if (isHidden && !isOwner) {
			ticketElement.style.display = "none";
		} else {
			ticketElement.style.display = "flex";
		}
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
