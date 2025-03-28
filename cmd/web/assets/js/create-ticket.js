function createClosePopverEvent() {
	document.dispatchEvent(new CloseModalEvent());
}

/** @type {HTMLFormElement | null} */
// @ts-ignore
const formElement = document.getElementById("ticket-form");
/** @type {HTMLFormElement | null} */
// @ts-ignore
const jiraFormElement = document.getElementById("jira-ticket-form");
/** @type {HTMLDivElement | null} */

// Close popover when successfully created a ticket
formElement.addEventListener("htmx:afterRequest", (event) => {
	// @ts-ignore
	resetFormOnSuccess(event.detail);
	document.dispatchEvent(new CloseModalEvent());
});
jiraFormElement.addEventListener("htmx:afterRequest", () => {
	document.dispatchEvent(new CloseModalEvent());

	/** @type {HTMLInputElement | null} **/
	const search = document.querySelector("input[hx-target='#search-result']");
	search.value = "";
});

htmx.on("afterSwap", createClosePopverEvent);

/**
 * @param {HtmxResponseInfo} event
 */
function resetFormOnSuccess(event) {
	if (event.xhr.status !== 200) {
		console.error("Request failed");
		return;
	}

	formElement.reset();
	jiraFormElement.reset();
}
