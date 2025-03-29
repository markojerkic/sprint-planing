/** @type {HTMLFormElement | null} */
// @ts-ignore
const formElement = document.getElementById("ticket-form");

document.addEventListener("createdTicket", () => {
	document.dispatchEvent(new CloseModalEvent());
	formElement.reset();
	resetJiraSearch();
});

function resetJiraSearch() {
	/** @type {HTMLInputElement | null} **/
	const jiraSearchInput = document.querySelector("input[id='jira-search']");
	if (!jiraSearchInput) {
		return;
	}
	jiraSearchInput.value = "";
}
