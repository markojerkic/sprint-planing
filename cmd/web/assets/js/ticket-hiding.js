document.addEventListener(
    "htmx:wsBeforeMessage",
    /** @param {CustomEvent} e */
    function (e) {
        /** @type {string} */
        const message = e.detail.message;

        const json = isJsonWebSocketMessage(message, "hideTicket");
        if (!json) {
            return;
        }

        const ticketId = json.data["ticketID"];
        const isHidden = json.data["isHidden"];

        if (!ticketId) {
            // Hide all tickets
            document.querySelectorAll("ui-flashing-div").forEach(
                /** @param {HTMLElement} element */
                (element) => {
                    const isOwner =
                        element.getAttribute("data-is-owner") === "true";
                    if (isHidden && !isOwner) {
                        element.setAttribute("data-is-hidden", "true");
                    } else {
                        element.setAttribute("data-is-hidden", "false");
                    }
                },
            );

            const isOwner =
                document
                    .querySelector("ui-flashing-div")
                    .getAttribute("data-is-owner") === "true";
            if (isOwner) {
                // Find all buttons with hx-post="/ticket/hide", switch classs btn-warn to btn-success
                // and change text to "Reveal"
                document
                    .querySelectorAll("button[hx-post='/ticket/hide']")
                    .forEach(
                        /** @param {HTMLButtonElement} element */
                        (element) => {
                            element.classList.remove("btn-sm-warn");
                            element.classList.add("btn-sm-success");
                            element.innerText = "Reveal";
                        },
                    );
            }

            return;
        }

        /** @type {import("./components/flashing-div").FlashingDiv} */
        const ticketElement = document.querySelector(
            `ui-flashing-div[data-ticket-id="${ticketId}"]`,
        );
        const isOwner = ticketElement.getAttribute("data-is-owner") === "true";

        if (isHidden && !isOwner) {
            ticketElement.setAttribute("data-is-hidden", "true");
        } else if (!isOwner) {
            ticketElement.setAttribute("data-is-hidden", "false");
            ticketElement.flash();
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
