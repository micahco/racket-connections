const method = document.getElementById("contact-method")
const input = document.getElementById("contact-value")

const updateInput = (event) => {
    switch (method.value) {
        case "1":
            input.type = "email"
            input.placeholder = "bennybeaver@gmail.com"
            break
        
        case "2": // phone
            input.type = "tel"
            input.placeholder = "541-737-1000 "
            break

        case "3": // other
            input.placeholder = "Discord, IRC, Etc., Etc."

    }
}

window.addEventListener("load", updateInput)
method.addEventListener("change", updateInput)
