/*
 * Open/close the HTML dialog element with an adjacent button.
 */

const dialog = document.querySelector("dialog")
const container = document.querySelector("dialog > div")
const showButton = document.querySelector("dialog + button")
const closeButton = document.querySelector("dialog button")

dialog.addEventListener("click", () => dialog.close())
container.addEventListener("click", (e) => e.stopPropagation())
showButton.addEventListener("click", () => dialog.showModal())
closeButton.addEventListener("click", () => dialog.close())
