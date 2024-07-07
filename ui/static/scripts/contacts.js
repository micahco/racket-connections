/*
 * Update the contact method's value placeholder depending
 * on the selected method.
 */
const method = document.getElementById("contact-method")
const input = document.getElementById("contact-value")

const updateInput = (event) => {
    switch (method.value) {
        case "email":
            input.type = "email"
            input.placeholder = "bennybeaver@gmail.com"
            break
        
        case "phone":
            input.type = "tel"
            input.placeholder = "541-737-1000 "
            break

        case "other":
            input.type = "text"
            input.placeholder = "Discord, IRC, Etc., Etc."
    }
}

window.addEventListener("load", updateInput)
method.addEventListener("change", updateInput)
