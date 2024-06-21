const deletePostForm = document.getElementById("delete-post");

if (deletePostForm) {
    deletePostForm.addEventListener("submit", (e) => {
        return confirm("Confirm delete")
    })
}
