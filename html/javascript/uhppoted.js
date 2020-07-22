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

function warning(msg) {
  let message = document.getElementById('message');
  
  if (message != null) {
      message.innerText = msg;
      message.classList.add("warning");
  } else {
      alert(msg);
  }
}



