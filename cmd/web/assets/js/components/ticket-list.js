/**
 * @typedef {Object} Ticket
 * @property {number} id
 * @property {string} name
 * @property {boolean} isHidden
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
        const searchInput = /** @type {HTMLInputElement|null} */ (document.getElementById("ticket-list-search"));
        const currentValue = searchInput?.value || "";
        const hadFocus = searchInput === document.activeElement;
        const cursorPosition = searchInput?.selectionStart || 0;

        const amIOwner = document.querySelector("[hx-post='/ticket/hide-all']") !== null;
        console.log("am i owner", amIOwner);
        console.log("tickets", this.#tickets.filter((ticket) => amIOwner || !ticket.isHidden));

        this.innerHTML = `
        <div class="fixed bottom-0 top-0 left-0 my-auto max-h-[80vh] max-w-28 bg-input-bg z-10 hover:max-w-fit ease-in-out transition-all duration-300 hidden lg:block">

            <input
                class="max-w-28 px-2 border border-gray-300"
                type="text"
                placeholder="Search..."
                id="ticket-list-search"
                value="${currentValue}"
            />

            <div class="scrollbar flex flex-col h-fit max-h-[80vh] gap-2 text-sm p-2 overflow-y-auto text-right"
                style="direction: rtl;"
            >
                ${this.#filteredTickets
                .map((ticket) => `<span class="cursor-pointer hover:underline p-1 rounded" data-ticket-id="${ticket.id}" onclick="document.querySelector('[data-ticket-id=&quot;${ticket.id}&quot;]:not(:hover)')?.scrollIntoView({behavior:'smooth',block:'center'})">${ticket.name}</span>`).join("")}
            </div>
        </div>
    `;
        if (hadFocus) {
            const newInput = /** @type {HTMLInputElement} */ (document.getElementById("ticket-list-search"));
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
            const searchInput = /** @type {HTMLInputElement|null} */ (document.getElementById("ticket-list-search"));
            const currentFilter = searchInput?.value || "";
            this.#applyFilter(currentFilter);
            this.render();
        });
    }


    #registerTicketListSearch() {
        const search = document.getElementById("ticket-list-search");
        if (!search) {
            console.error("Ticket list search not found");
            return;
        }
        search.addEventListener("input", (event) => this.#onFilter(event));
    }

    #applyFilter(filterText = "") {
        const amIOwner = document.querySelector("[hx-post='/ticket/hide-all']") !== null;
        this.#filteredTickets = this.#tickets
            .filter(ticket => amIOwner || !ticket.isHidden)
            .filter(ticket => ticket.name.toLowerCase().includes(filterText.toLowerCase()));
    }

    /** @param {Event} event */
    #onFilter(event) {
        const target = /** @type {HTMLInputElement} */ (event.target);
        console.log("filter", target.value);
        if (this.#debounceTimer) {
            clearTimeout(this.#debounceTimer);
        }

        this.#debounceTimer = setTimeout(() => {
            const filter = target.value;
            console.log("filter", filter);
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
        console.log("tickets", this.#tickets);
        this.render();
    }
}
customElements.define("ui-ticket-list", TicketListElement);
