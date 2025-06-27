const form = document.querySelector("form");
eField = form.querySelector(".email"),
    eInput = eField.querySelector("input"),
    pField = form.querySelector(".password"),
    pInput = pField.querySelector("input");



eInput.addEventListener("input", hideLoginError);
eInput.addEventListener("input", checkEmail);
pInput.addEventListener("input", hideLoginError);
pInput.addEventListener("input", checkPass);

form.onsubmit = (e) => {
    e.preventDefault();

    (eInput.value == "") ? eField.classList.add("shake", "error"): checkEmail();
    (pInput.value == "") ? pField.classList.add("shake", "error"): checkPass();

    setTimeout(() => {
        eField.classList.remove("shake");
        pField.classList.remove("shake");
    }, 500);


    if (!eField.classList.contains("error") && !pField.classList.contains("error")) {
        submitForm(eInput.value, pInput.value);
    }
}

function checkEmail() {
    let pattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!eInput.value.match(pattern)) {
        eField.classList.add("error");
        eField.classList.remove("valid");
        let errorTxt = eField.querySelector(".error-txt");

        (eInput.value != "") ? errorTxt.innerText = "Enter a valid email address": errorTxt.innerText = "Email can't be blank";
    } else {
        eField.classList.remove("error");
        eField.classList.add("valid");
    }
}

function checkPass() {
    if (pInput.value == "") {
        pField.classList.add("error");
        pField.classList.remove("valid");
    } else {
        pField.classList.remove("error");
        pField.classList.add("valid");
    }
}

function loginError(msg) {
    const loginField = document.querySelector(".login-error");
    const errorTxt = loginField.querySelector(".error-txt");
    errorTxt.innerText = msg;
    loginField.classList.add("error");
}

function hideLoginError() {
    const loginField = document.querySelector(".login-error");
    loginField.classList.remove("error");
}

async function submitForm(email, password) {
    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username: email, password: password })
        });

        if (!response.ok) {
            const error = await response.text();
            loginError("Login failed: " + error);
            return;
        }

        const result = await response.json();
        alert("Login successful! Token: " + result.token);

        // Пример редиректа после входа:
        window.location.href = "/dashboard"; // или другая защищённая страница
    } catch (err) {
        console.error("Network error:", err);
        loginError("Network error. Try again later.");
    }
}

