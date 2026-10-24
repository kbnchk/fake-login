(function () {
    "use strict";
    var $ = function (id) { return document.getElementById(id); };
    var emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

    var year = $("year");
    if (year) year.textContent = new Date().getFullYear();

    function setInvalid(field, bad) {
        if (!field) return;
        field.classList.toggle("invalid", bad);
        if (bad) {
            field.classList.remove("shake");
            void field.offsetWidth;
            field.classList.add("shake");
        }
    }
    function showAlert(text) {
        var a = $("alert"), t = $("alert-text");
        if (!a || !t) return;
        t.textContent = text;
        a.classList.add("show");
    }
    function hideAlert() { var a = $("alert"); if (a) a.classList.remove("show"); }

    // ---------- страница входа ----------
    var login = $("login");
    if (login) {
        var email = $("email"), pass = $("password"), btn = $("submit");

        email.addEventListener("input", function () { setInvalid($("f-email"), false); hideAlert(); });
        pass.addEventListener("input", function () { setInvalid($("f-pass"), false); hideAlert(); });

        $("toggle").addEventListener("click", function () {
            var show = pass.type === "password";
            pass.type = show ? "text" : "password";
            this.setAttribute("aria-pressed", String(show));
            this.setAttribute("aria-label", show ? "Hide password" : "Show password");
        });

        $("sso").addEventListener("click", function () {
            showAlert("Single sign-on is not configured for this workspace.");
        });
        var ra = $("request-access");
        if (ra) ra.addEventListener("click", function (e) {
            e.preventDefault();
            showAlert("Access requests are handled by your workspace administrator.");
        });

        login.addEventListener("submit", function (e) {
            e.preventDefault();
            var badEmail = !emailRe.test(email.value.trim());
            var badPass = pass.value === "";
            setInvalid($("f-email"), badEmail);
            setInvalid($("f-pass"), badPass);
            if (badEmail || badPass) return;

            btn.disabled = true;
            btn.classList.add("busy");
            $("submit-text").textContent = "Signing in…";
            fetch("/api/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username: email.value.trim(), password: pass.value })
            }).then(function (r) {
                return r.json().catch(function () { return {}; }).then(function (j) {
                    showAlert(j.message || "Something went wrong. Please try again.");
                });
            }).catch(function () {
                showAlert("Network error. Please try again later.");
            }).then(function () {
                btn.disabled = false;
                btn.classList.remove("busy");
                $("submit-text").textContent = "Sign in";
                pass.value = "";
                pass.focus();
            });
        });
    }

    // ---------- сброс пароля ----------
    var forgot = $("forgot");
    if (forgot) {
        forgot.addEventListener("submit", function (e) {
            e.preventDefault();
            var bad = !emailRe.test($("email").value.trim());
            setInvalid($("f-email"), bad);
            if (bad) return;
            $("alert").classList.add("show");
            forgot.querySelector("button").disabled = true;
        });
    }
})();
