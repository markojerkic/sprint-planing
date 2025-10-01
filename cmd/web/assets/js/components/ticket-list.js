/**
 * @typedef {Object} Ticket
 * @property {number} id
 * @property {string} name
 * @property {boolean} isHidden
 * @property {boolean} isClosed
 */
class TicketListElement extends HTMLElement {
    /** @type {Ticket[]} */
    #tickets = [];

    /** @type {Ticket[]} */
    #filteredTickets = [];

    /** @type {NodeJS.Timeout|null} */
    #debounceTimer;

    connectedCallback() {
        this.render();
        this.#fetchTickets();
        this.#registerTicketListSearch();
        this.#registerRefreshTickets();
    }

    render() {
        const searchInput = /** @type {HTMLInputElement|null} */ (
            document.getElementById("ticket-list-search")
        );
        const currentValue = searchInput?.value || "";
        const hadFocus = searchInput === document.activeElement;
        const cursorPosition = searchInput?.selectionStart || 0;

        this.innerHTML = `
        <div class="fixed bottom-0 top-0 left-0 my-auto max-h-[80vh] max-w-28 bg-input-bg z-10 hover:max-w-fit ease-in-out transition-all duration-300 hidden lg:flex lg:flex-col">

            <input
                class="max-w-28 px-2 border border-gray-300"
                type="text"
                placeholder="Search..."
                id="ticket-list-search"
                value="${currentValue}"
            />

            <div class="scrollbar flex flex-col h-fit max-h-[80vh] gap-2 text-sm p-2 overflow-y-auto text-right flex-grow"
                style="direction: rtl;"
            >
                ${this.#filteredTickets.map((ticket) => `<ui-ticket-list-item data-ticket='${JSON.stringify(ticket)}'></ui-ticket-list-item>`).join("")}
            </div>
            <ul class="gap-2 list-disc pl-6 sticky mt-auto bottom-0 left-0 right-0 bg-input-bg z-10 p-2">
                <li class="text-green-300">Revealed</li>
                <li class="text-gray-400">Hidden</li>
                <li class="text-red-300">Closed</li>
            </ul>
        </div>
    `;
        if (hadFocus) {
            const newInput = /** @type {HTMLInputElement} */ (
                document.getElementById("ticket-list-search")
            );
            newInput.focus();
            newInput.setSelectionRange(cursorPosition, cursorPosition);
        }
    }

    #registerRefreshTickets() {
        htmx.on("htmx:wsBeforeMessage", (e) => {
            // this.#fetchTickets();
            const message = e.detail.message;
            const json = isJsonWebSocketMessage(message, "refreshTicketList");
            if (!json) {
                return;
            }
            console.log("refreshing tickets", json);
            this.#tickets = json.data;
            const searchInput = /** @type {HTMLInputElement|null} */ (
                document.getElementById("ticket-list-search")
            );
            const currentFilter = searchInput?.value || "";
            this.#applyFilter(currentFilter);
            this.render();
        });
    }

    #registerTicketListSearch() {
        this.addEventListener("input", (event) => {
            const target = /** @type {HTMLElement} */ (event.target);
            if (target.id === "ticket-list-search") {
                this.#onFilter(event);
            }
        });
    }

    #applyFilter(filterText = "") {
        const amIOwner =
            document.querySelector("[hx-post='/ticket/hide-all']") !== null;
        this.#filteredTickets = this.#tickets
            .filter((ticket) => amIOwner || !ticket.isHidden)
            .filter((ticket) =>
                ticket.name.toLowerCase().includes(filterText.toLowerCase()),
            );
    }

    /** @param {Event} event */
    #onFilter(event) {
        const target = /** @type {HTMLInputElement} */ (event.target);
        if (this.#debounceTimer) {
            clearTimeout(this.#debounceTimer);
        }

        this.#debounceTimer = setTimeout(() => {
            const filter = target.value;
            this.#applyFilter(filter);
            this.render();
            this.#debounceTimer = null;
        }, 500);
    }

    async #fetchTickets() {
        this.#tickets = await fetch(window.location.href, {
            method: "GET",
            headers: {
                Accept: "application/json",
            },
        }).then((response) => {
            if (!response.ok) {
                throw new Error("Failed to fetch tickets");
            }
            return response.json();
        });

        this.#applyFilter();
        this.render();
    }
}

class TicketListItemElement extends HTMLElement {
    /** @type {Ticket} */
    ticket;

    connectedCallback() {
        const ticketData = this.getAttribute("data-ticket");
        this.ticket = ticketData ? JSON.parse(ticketData) : null;

        if (!this.ticket) return;

        this.innerHTML = `
                <span class="cursor-pointer hover:underline p-1 rounded ${this.#getItemTextColor(this.ticket)}" data-ticket-id="${this.ticket.id}" onclick="document.querySelector('ui-flashing-div[data-ticket-id=&quot;${this.ticket.id}&quot;]')?.flash()">${this.ticket.name}</span>
        `;
    }

    /**
     * @param {Ticket} ticket
     * @returns {string}
     */
    #getItemTextColor(ticket) {
        if (ticket.isClosed) {
            return "text-red-300";
        }
        if (ticket.isHidden) {
            return "text-gray-400";
        }
        return "text-green-300";
    }
}

customElements.define("ui-ticket-list", TicketListElement);
customElements.define("ui-ticket-list-item", TicketListItemElement);
