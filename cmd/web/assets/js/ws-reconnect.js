let wsHasConnectedBefore = false;
let isRefetching = false;

htmx.on("htmx:wsOpen", function() {
    if (!wsHasConnectedBefore) {
        wsHasConnectedBefore = true;
        return;
    }

    if (isRefetching) {
        return;
    }

    console.log('Reconnecting...');
    htmx.ajax('get', window.location.pathname, {
        target: '#ticket-list',
        select: '#ticket-list',
        // swap: 'outerHTML'
    }).then(function() {
        isRefetching = false;
    });
})
