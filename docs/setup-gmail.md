# 🚀 Quick Gmail SMTP Setup (5 Minutes)

Follow these steps to enable email sending to ANY address using your Gmail account:

## Step 1: Enable 2-Factor Authentication

1. Go to: https://myaccount.google.com/security
2. Scroll to "2-Step Verification"
3. Click "Get Started" and follow the prompts
4. ✅ You should see "2-Step Verification is on"

## Step 2: Create App Password

1. Go to: https://myaccount.google.com/apppasswords
   - If you don't see this option, make sure 2FA is enabled first
2. In the "Select app" dropdown, choose **"Mail"**
3. In the "Select device" dropdown, choose **"Other (Custom name)"**
4. Type: **"NetLink Backend"** or any name you want
5. Click **"Generate"**
6. You'll see a 16-character password like: `abcd efgh ijkl mnop`
7. **COPY THIS PASSWORD** (you won't see it again!)

## Step 3: Update Your .env File

Open `go-backend/.env` and update it:

```env
# Database
DATABASE_URL=<db-url>

# App
APP_URL=http://localhost:3000

# Gmail SMTP (NEW - Add these lines)
GMAIL_USERNAME=<email>
GMAIL_APP_PASSWORD=<app-password>
FROM_EMAIL=NetLink <m.wasifmalik17@gmail.com>

# Resend (Keep for backup, but Gmail will be used if configured)
RESEND_API_KEY=<api_key>  
```

**Important:** 
- Replace `abcdefghijklmnop` with your actual 16-character App Password
- Remove any spaces from the App Password
- You can use your email address as-is

## Step 4: Restart Your Server

```bash
# Stop the current server (Ctrl+C)

# Rebuild
go build -o server.exe

# Start
./server.exe
```

## Step 5: Verify It's Working

When your server starts, you should see:
```
📧 Email service initialized with Gmail SMTP
```

Now try registering a new user with ANY email address - it should work! 🎉

## 🧪 Test Email Sending

You can test by:
1. Registering a new account with any email
2. Using the forgot password feature
3. Or use this curl command:

```bash
curl -X POST http://localhost:8080/api/test-email \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"any-email@example.com\"}"
```

## ⚠️ Troubleshooting

**"Username and Password not accepted":**
- Make sure you're using the App Password, not your regular Gmail password
- Remove any spaces from the App Password
- Make sure 2FA is enabled on your Gmail account

**Still seeing "onboarding@resend.dev":**
- Check that `GMAIL_USERNAME` and `GMAIL_APP_PASSWORD` are set in .env
- Restart your server
- Check server logs for "Email service initialized with Gmail SMTP"

**Emails going to spam:**
- This is normal for the first few emails
- Ask recipients to mark as "Not Spam"
- After a few emails, Gmail will learn and deliver to inbox

## 📊 Gmail SMTP Limits

- **500 emails per day** (plenty for most apps)
- **Completely free**
- **Works with any email address**

That's it! You're all set up! 🚀
