var registerButton = document.getElementById('signup-button');


function register() {
    name = document.getElementById('name').value;
    email = document.getElementById('email').value;
    password = document.getElementById('password').value;

    data = {
        "Name": name,
        "Email": email,
        "Password": password
    }

    axios.post('/register', data)
        .then(function(response) {
            var message = document.createElement('h3')
            if(response.data['Error'] == null) {
                message.innerText = 'Successfully registered!';
                document.getElementById('register-window').appendChild(message)
            } else {
                message.innerText = response.data['Error'].Message
                document.getElementById('register-window').appendChild(message)
            }
            console.log(response)
        })
        .catch(function(error) {
            console.log(error)
        })
}

registerButton.addEventListener('click', function(event){
    register();
    event.preventDefault();
})