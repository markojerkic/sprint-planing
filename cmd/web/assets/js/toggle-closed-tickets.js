function toggleClosedTickets() {
	const isHidden = localStorage.getItem("hideClosedTickets") === "true";

	document.body.classList.toggle("hide-closed-tickets");

	localStorage.setItem("hideClosedTickets", !isHidden ? "true" : "false");
	setCurrentState();
}

function setCurrentState() {
	const isHidden = localStorage.getItem("hideClosedTickets") === "true";
	const toggleButton = document.getElementById("toggle-hidden-tickets");
	if (!toggleButton) {
		console.error("Toggle button not found");
		return;
	}
	if (isHidden) {
		toggleButton.textContent = "Show Closed Tickets";
		toggleButton.classList.add("btn-sm-success");
		toggleButton.classList.remove("btn-sm-warning");
	} else {
		toggleButton.textContent = "Hide Closed Tickets";
		toggleButton.classList.add("btn-sm-warning");
		toggleButton.classList.remove("btn-sm-success");
	}
}

document.addEventListener("DOMContentLoaded", setCurrentState);
document.addEventListener("htmx:afterSwap", (e) => {
	const isBoosted = e.detail.requestConfig?.boosted;
	if (isBoosted) {
		setCurrentState();
	}
});
