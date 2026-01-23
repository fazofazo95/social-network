const form = document.getElementById('signup');
const output = document.getElementById('output');

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  const data = {
    username: document.getElementById('username').value,
    email: document.getElementById('email').value,
    password: document.getElementById('password').value,
  };

  output.textContent = 'Sending...';

  try {
    const res = await fetch('/api/signup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    const text = await (async () => {
      try { return await res.text(); } catch (e) { return String(e); }
    })();

    output.textContent = `Status: ${res.status} ${res.statusText}\n\n${text}`;
  } catch (err) {
    output.textContent = `Request error:\n${err}`;
  }
});