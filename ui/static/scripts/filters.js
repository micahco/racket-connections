const form = document.getElementById("filters")
form.addEventListener("change", function() {
    form.submit()
});

form.querySelector("button[type='submit']").classList.add("hidden")

const tableBtns = form.querySelectorAll("table button")
for (const btn of tableBtns) {
    btn.addEventListener("click", () => {
        form.submit()
    })
}
