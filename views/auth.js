var loginButton = document.getElementById('login-button');

function login() {
    username = document.getElementById('login').value;
    password = document.getElementById('password').value;

    data = {
        "Email": username,
        "Password": password
    }

    axios.post('/login', data)
        .then(function(response) {
            var message = document.createElement('h3')
            message.innerText = response.data.message
            document.getElementById('login-window').appendChild(message)

            localStorage.setItem('token', response.data.token)
            window.location.href = 'auth/vd'
        })
        .catch(function (error) {
            console.log(error)
        });
}

loginButton.addEventListener('click', function(event){
    login();
    event.preventDefault();
})