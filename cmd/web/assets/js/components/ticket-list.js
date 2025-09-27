/**
 * @typedef {Object} Ticket
 * @property {number} id
 * @property {string} name
 * @property {boolean} isClosed
 */
class TicketListElement extends HTMLElement {

    /** @type {Ticket[]} */
    #tickets = [];

    connectedCallback() {
        this.render();
        this.#fetchTickets();
    }

    render() {
        this.innerHTML = `
        <div class="fixed bottom-0 top-0 left-0 my-auto h-[80vh] max-w-28 bg-input-bg z-10 hover:max-w-fit ease-in-out transition-all duration-300">
            <div class="scrollbar flex flex-col max-h-full gap-2 text-sm p-2 overflow-y-auto text-right"
                style="direction: rtl;"
            >
                ${this.#tickets.map((ticket) => `<span class="cursor-pointer hover:underline p-1 rounded" data-ticket-id="${ticket.id}" onclick="document.querySelector('[data-ticket-id=&quot;${ticket.id}&quot;]:not(:hover)')?.scrollIntoView({behavior:'smooth',block:'center'})">${ticket.name}</span>`).join("")}
            </div>
        </div>
    `;
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

        console.log("tickets", this.#tickets);
        this.render();
    }
}
customElements.define("ui-ticket-list", TicketListElement);
