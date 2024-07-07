/*
 * Checks the initial state of each checkbox in the filters form.
 * Listens for any change events and then compares the new state 
 * to the initial state. If there are any changes, apply the 
 * ".highlight" classs to the submit button.
 */

const form = document.getElementById("filters")
const submit = form.querySelector("button[type='submit']")
const checkboxes = form.querySelectorAll("input[type='checkbox']")

const initialState = []

for (let i = 0; i < checkboxes.length; i++) {
    initialState[i] = checkboxes[i].checked
}

checkboxes.forEach((el) => {
    el.addEventListener("change", (e) => {
        for (let i = 0; i < checkboxes.length; i++) {
            if (checkboxes[i].checked != initialState[i]) {
                submit.classList.add("highlight")
                return
            }
        }

        submit.classList.remove("highlight")
    })
})

const available = form.querySelector("a[href='/posts/available']")
available.href += window.location.search
