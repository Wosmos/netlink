# 🚀 Quick Start: Enable Email for All Users

## Current Problem
- ✅ Emails work for: `m.wasifmalik17@gmail.com`
- ❌ Emails DON'T work for: other email addresses
- 📧 Shows as from: `onboarding@resend.dev`

## Solution: Use Gmail SMTP (5 Minutes Setup)

### What You'll Get:
- ✅ Send emails to **ANY email address**
- ✅ Emails show from: **NetLink <m.wasifmalik17@gmail.com>**
- ✅ **Completely FREE** (500 emails/day)
- ✅ **No domain required**

---

## 📋 Setup Steps

### 1️⃣ Get Gmail App Password

1. **Enable 2FA on Gmail:**
   - Go to: https://myaccount.google.com/security
   - Enable "2-Step Verification"

2. **Create App Password:**
   - Go to: https://myaccount.google.com/apppasswords
   - Select: Mail → Other (Custom name)
   - Name it: "NetLink Backend"
   - Click "Generate"
   - **COPY the 16-character password** (e.g., `abcd efgh ijkl mnop`)

### 2️⃣ Update .env File

Open `go-backend/.env` and add these lines:

```env
GMAIL_USERNAME=m.wasifmalik17@gmail.com
GMAIL_APP_PASSWORD=paste_your_16_char_password_here
```

**Example:**
```env
GMAIL_USERNAME=m.wasifmalik17@gmail.com
GMAIL_APP_PASSWORD=abcdefghijklmnop
FROM_EMAIL=NetLink <m.wasifmalik17@gmail.com>
```

### 3️⃣ Restart Server

```bash
# Stop server (Ctrl+C)
cd go-backend
go build -o server.exe
./server.exe
```

Look for this message:
```
📧 Email service initialized with Gmail SMTP
```

### 4️⃣ Test It!

Register a new user with **any email address** - it should work now! 🎉

---

## 🎯 Alternative: Custom Domain (More Professional)

If you want emails from your own domain (e.g., `noreply@yourapp.com`):

1. **Buy a domain** (~$10/year from Namecheap, GoDaddy, etc.)
2. **Add to Resend:** https://resend.com/domains
3. **Add DNS records** (provided by Resend)
4. **Update .env:**
   ```env
   FROM_EMAIL=noreply@yourapp.com
   ```

See `EMAIL_SETUP_GUIDE.md` for detailed instructions.

---

## 📚 Documentation

- **`setup-gmail.md`** - Detailed Gmail SMTP setup with screenshots
- **`EMAIL_SETUP_GUIDE.md`** - Complete guide for all email options
- **Current setup** - Check your `.env` file

---

## ❓ Need Help?

**Emails not sending?**
- Check `.env` has `GMAIL_USERNAME` and `GMAIL_APP_PASSWORD`
- Make sure App Password has no spaces
- Restart your server

**Still showing onboarding@resend.dev?**
- Gmail SMTP might not be configured
- Check server logs for "Email service initialized with Gmail SMTP"

**Emails going to spam?**
- Normal for first few emails
- Ask recipients to mark as "Not Spam"
- Gmail will learn and deliver to inbox after a few emails

---

That's it! Follow the 3 steps above and you'll be sending emails to anyone! 🚀
