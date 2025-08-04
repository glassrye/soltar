# Soltar VPN Registration Webapp

A simple, beautiful web interface for client registration and OTP verification.

## Features

- **Clean, modern UI** with gradient backgrounds
- **Email registration** with OTP verification
- **Responsive design** that works on all devices
- **Real-time feedback** with success/error messages
- **Secure OTP input** with auto-focus and validation
- **Client information storage** in localStorage

## Usage

### Local Development

1. Start the Go server:
```bash
cd cmd/worker
go run main.go
```

2. Serve the webapp:
```bash
cd webapp
python3 -m http.server 8000
# or
npx serve .
```

3. Visit `http://localhost:8000`

### Production Deployment

The webapp is served by the Go server at the root path. Simply deploy the Go server and the registration form will be available at the root URL.

## API Integration

The webapp integrates with the Soltar VPN API:

- `POST /register` - Register with email
- `POST /verify` - Verify OTP and get client credentials

## Client Deployment

Each client gets their own isolated deployment:

```bash
# Deploy a new client instance
./deploy-client.sh john@example.com john-vpn

# This creates:
# - App: soltar-john-vpn-1234567890
# - Volume: soltar-john-vpn-data
# - URL: https://soltar-john-vpn-1234567890.fly.dev
```

## Customization

### Styling
The webapp uses CSS custom properties for easy theming:

```css
:root {
  --primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  --border-color: #e1e5e9;
  --focus-color: #667eea;
}
```

### Branding
Update the logo and title in `index.html`:

```html
<div class="logo">
    <h1>🔒 Your Brand VPN</h1>
</div>
```

## Security Features

- **HTTPS only** in production
- **CORS headers** for cross-origin requests
- **Input validation** on both client and server
- **Secure OTP handling** with timeouts
- **No sensitive data** stored in localStorage (only client ID and token)

## Browser Support

- Chrome/Edge 88+
- Firefox 85+
- Safari 14+
- Mobile browsers (iOS Safari, Chrome Mobile)

## Development

### Adding Features

1. **New fields**: Add to the registration form and update the API calls
2. **Custom validation**: Add JavaScript validation functions
3. **Additional steps**: Extend the OTP verification flow
4. **Styling**: Modify the CSS in the `<style>` tag

### Testing

```bash
# Test the registration flow
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Test OTP verification
curl -X POST http://localhost:8080/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456"}'
```

## Deployment Checklist

- [ ] Configure email service for OTP delivery
- [ ] Set up custom domain (optional)
- [ ] Update JWT secret in production
- [ ] Test registration flow end-to-end
- [ ] Verify health check endpoint
- [ ] Monitor logs and metrics 