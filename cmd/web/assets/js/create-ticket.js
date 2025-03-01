// @ts-check
/// <reference path="./htmx.d.ts" />

/** @type {HTMLFormElement | null} */
const formElement = document.getElementById("ticket-form");
/** @type {HTMLDivElement | null} */
const popoverElement = document.getElementById("create-ticket-popover");

popoverElement.addEventListener("keydown", function (event) {
  if (event.key === "Escape") {
    console.log("Escape key pressed");
    popoverElement.hidePopover();
  }
});

popoverElement.addEventListener("toggle", function (event) {
  if (event.newState === "closed") {
    formElement.reset();
  }
  if (event.newState === "open") {
    formElement.querySelector("input[name='ticketName']").focus();
  }
});

// Close popover when successfully created a ticket
formElement.addEventListener(
  "htmx:afterRequest",
  (event) => resetFormOnSuccess(event.detail),
  popoverElement.hidePopover(),
);

/**
 * @param {HtmxResponseInfo} event
 */
function resetFormOnSuccess(event) {
  if (event.xhr.status !== 200) {
    console.error("Request failed");
    return;
  }

  formElement.reset();
  document.getElementById("create-ticket-popover").hidePopover();
}
