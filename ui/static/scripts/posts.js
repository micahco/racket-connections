const getSiblingsAfter = (node) => {
    const siblings = []
    let cur = node
    while (cur.nextElementSibling) {
        cur = cur.nextElementSibling
        siblings.push(cur)
    }

    return siblings
}

const checkBox = (els) => {
    let allChecked = true
    for (const td of els) {
        if (!td.querySelector("input").checked) {
            allChecked = false
            break
        }
    }

    for (const td of els) {
        td.querySelector("input").checked = allChecked ? false : true
    }
}



for (const th of document.querySelectorAll(".day-header")) {
    th.addEventListener("click", e => {
        checkBox(getSiblingsAfter(th))
    })
}
