const state = {
    searchButton: document.getElementById("searchButton"),
    searchInput: document.getElementById("searchInput"),
    responseBody: document.getElementById("responseBody"),
    rfcHiddenInput: document.getElementById("rfc"),
    loader: document.getElementById("loader"),
    getMarkDownContainer: () => {
        return document.getElementById("markdown");
    },
    getMarkDown2Container: () => {
        return document.getElementById("markdown2");
    }
}
function initMd() {
    // Renderizar en el div
    const container = state.getMarkDownContainer();
    if (!container) return;
    const content = state.getMarkDownContainer().innerHTML;
    state.getMarkDownContainer().innerHTML = (marked.parse(content))

    const container2 = state.getMarkDown2Container();
    if (!container2) return;
    const content2 = state.getMarkDown2Container().innerHTML;
    state.getMarkDown2Container().innerHTML = (marked.parse(content2))
}
async function buscar(rfc) {
    try {
        const response = await fetch("/buscar?rfc=" + rfc);
        const html = await response.text();
        return html;
    }
    catch (ex) {
        return `<h3>Error<h3>`
    }
}
async function onSearch(rfc) {

    state.responseBody.innerHTML = ""
    console.log("Search")
    console.log("value: ", rfc);
    if (rfc === "") {
        return
    }
    const responseHtml = await buscar(rfc);
    state.responseBody.innerHTML = responseHtml
    initMd();
}
async function main() {
    state.searchButton.onclick = async () => {
        state.loader.style.display = "block"
        const value = state.searchInput.value;
        await onSearch(value);
        state.loader.style.display = "none"
    }

    if (state.rfcHiddenInput.value !== "" || state.rfcHiddenInput.value !== undefined) {
        state.loader.style.display = "block"
        state.searchInput.value = state.rfcHiddenInput.value;
        await onSearch(state.rfcHiddenInput.value);
        state.loader.style.display = "none"
    }
}

main().catch(console.error);
