var idleTimer;

document.addEventListener('mousedown', event => {
  resetIdle(event);
});

document.addEventListener('click', event => {
  resetIdle(event);
});

document.addEventListener('scroll', event => {
  resetIdle(event);
});

document.addEventListener('keypress', event => {
  resetIdle(event);
});

async function postAsForm(url='', data={}) {
  let pairs = [];
  for (name in data) {
    pairs.push(encodeURIComponent(name) + '=' + encodeURIComponent(data[name]));
  }

  const response = await fetch(url, { method: 'POST', 
                                      mode: 'cors',
                                      cache: 'no-cache',
                                      credentials: 'same-origin',
                                      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                                      redirect: 'follow', 
                                      referrerPolicy: 'no-referrer', 
                                      body: pairs.join('&').replace( /%20/g,'+')
                                    });
  return response; 
}

async function postAsJSON(url='', data={}) {
  const response = await fetch(url, { method: 'POST', 
                                      mode: 'cors',
                                      cache: 'no-cache',
                                      credentials: 'same-origin',
                                      headers: { 'Content-Type': 'application/json' },
                                      redirect: 'follow', 
                                      referrerPolicy: 'no-referrer', 
                                      body: JSON.stringify(data)
                                    });
  return response; 
}

function logout(event) {
  if (event != null) {
     event.preventDefault();    
  }

  postAsJSON('/logout', {})
    .then(response => { 
        if (response.status == 200 && response.redirected) {
           window.location = response.url;
        } else {
           return response.text()
        }
    })
    .then(msg => { 
        warning(msg);
    })
    .catch(function(err) { console.error(err) });
}

function warning(msg) {
  let message = document.getElementById('message');
  
  if (message != null) {
      message.innerText = msg;
      message.classList.add("warning");
  } else {
      alert(msg);
  }
}

function onIdle() {
  logout();
}

function resetIdle() {
  if (idleTimer != null) {
      clearTimeout(idleTimer);
  }
  
  idleTimer = setTimeout(onIdle, 15*60*1000);    
}

