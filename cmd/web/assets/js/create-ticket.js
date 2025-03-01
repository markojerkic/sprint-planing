// @ts-check
/// <reference path="./htmx.d.ts" />

/** @type {HTMLFormElement | null} */
const formElement = document.getElementById("ticket-form");
/** @type {HTMLDivElement | null} */
const popoverElement = document.getElementById("create-ticket-popover");

popoverElement.addEventListener("keydown", function (event) {
  if (event.key === "Escape") {
    console.log("Escape key pressed");
    document.getElementById("create-ticket-popover").hidePopover();
  }
});

// Close popover when successfully created a ticket
formElement.addEventListener("htmx:afterRequest", (event) =>
  resetFormOnSuccess(event.detail),
);

/**
 * @param {HtmxResponseInfo} event
 */
function resetFormOnSuccess(event) {
  console.log("afterRequest", formElement);
  formElement.reset();
  document.getElementById("create-ticket-popover").hidePopover();
}
