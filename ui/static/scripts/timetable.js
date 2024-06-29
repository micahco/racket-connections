const toggleChecked = (nodes) => {
    let allChecked = true
    for (const td of nodes) {
        if (!td.querySelector("input").checked) {
            allChecked = false
            break
        }
    }

    for (const td of nodes) {
        td.querySelector("input").checked = allChecked ? false : true
    }
}

const rowNodes = (th) => {
    const nodes = []

    let cur = th
    while (cur.nextElementSibling) {
        cur = cur.nextElementSibling
        nodes.push(cur)
    }

    return nodes
}

const table = document.getElementById("timetable")

const colNodes = (col) => {
    const nodes = []

    const rows = table.querySelector("tbody").getElementsByTagName("tr")
    for (const row of rows) {
        const td = row.getElementsByTagName("td")[col]
        nodes.push(td)
    }

    return nodes
}

for (const th of table.querySelectorAll("th[scope='row']")) {
    th.querySelector("button").addEventListener("click", e => {
        toggleChecked(rowNodes(th))
    })
}

const colThs = table.querySelectorAll("th[scope='col']")
for (let i = 0; i < colThs.length; i++) {
    colThs[i].querySelector("button").addEventListener("click", e => {
        toggleChecked(colNodes(i))
    })
}
