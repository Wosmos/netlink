# 📧 Email Setup Guide

This guide explains how to set up email delivery for your application with multiple options.

## 🎯 Option 1: Resend with Custom Domain (Recommended for Production)

### Benefits:
- ✅ Send to ANY email address
- ✅ Custom "from" email (e.g., noreply@yourdomain.com)
- ✅ Professional appearance
- ✅ Better deliverability
- ✅ Free tier: 3,000 emails/month, 100 emails/day

### Setup Steps:

#### 1. Get a Domain
You need a domain name (e.g., yourapp.com). You can buy one from:
- Namecheap (~$10/year)
- GoDaddy
- Cloudflare
- Google Domains

#### 2. Add Domain to Resend
1. Go to https://resend.com/domains
2. Click "Add Domain"
3. Enter your domain (e.g., `yourapp.com`)
4. Resend will show you DNS records to add

#### 3. Add DNS Records
Go to your domain provider's DNS settings and add these records:

**SPF Record:**
```
Type: TXT
Name: @
Value: v=spf1 include:amazonses.com ~all
```

**DKIM Record:**
```
Type: TXT
Name: resend._domainkey
Value: [provided by Resend - long string]
```

**Return-Path (optional but recommended):**
```
Type: CNAME
Name: resend
Value: feedback-smtp.us-east-1.amazonses.com
```

#### 4. Wait for Verification
- DNS propagation takes 5-60 minutes
- Resend will automatically verify once DNS is updated
- You'll see a green checkmark when verified

#### 5. Update Your .env
```env
RESEND_API_KEY=re_your_api_key_here
FROM_EMAIL=noreply@yourdomain.com
# Or with a custom name:
# FROM_EMAIL=YourApp <noreply@yourdomain.com>
```

#### 6. Restart Your Server
```bash
# Stop the server (Ctrl+C)
# Rebuild and start
go build -o server.exe
./server.exe
```

---

## 🎯 Option 2: Gmail SMTP (Free, No Domain Needed)

### Benefits:
- ✅ Completely free
- ✅ No domain required
- ✅ Send to any email address
- ✅ Uses your Gmail account

### Limitations:
- ⚠️ 500 emails per day limit
- ⚠️ Shows "via gmail.com" in recipient's inbox
- ⚠️ Less professional for production

### Setup Steps:

#### 1. Enable 2-Factor Authentication on Gmail
1. Go to https://myaccount.google.com/security
2. Enable "2-Step Verification"

#### 2. Create App Password
1. Go to https://myaccount.google.com/apppasswords
2. Select "Mail" and "Other (Custom name)"
3. Name it "YourApp Backend"
4. Click "Generate"
5. Copy the 16-character password (e.g., `abcd efgh ijkl mnop`)

#### 3. Update Your .env
```env
# Comment out or remove Resend settings
# RESEND_API_KEY=...
# FROM_EMAIL=...

# Add Gmail SMTP settings
GMAIL_USERNAME=your.email@gmail.com
GMAIL_APP_PASSWORD=abcdefghijklmnop
FROM_EMAIL=your.email@gmail.com
```

#### 4. Restart Your Server
```bash
go build -o server.exe
./server.exe
```

You should see: `📧 Email service initialized with Gmail SMTP`

---

## 🎯 Option 3: Resend Free Tier (Current Setup)

### Current Limitations:
- ❌ Only sends to verified email addresses
- ❌ Must use onboarding@resend.dev
- ✅ Good for testing

### To Add More Recipients:
1. Go to https://resend.com/settings/team
2. Invite team members with their email addresses
3. They'll receive verification emails
4. Once verified, you can send to those addresses

---

## 🔍 Troubleshooting

### Emails Not Arriving?

**Check Spam Folder:**
- Gmail, Outlook, etc. may filter emails to spam initially
- Mark as "Not Spam" to improve future deliverability

**Verify Configuration:**
```bash
# Check your .env file
cat .env | grep -E "RESEND|GMAIL|FROM_EMAIL"
```

**Check Server Logs:**
Look for these messages when sending emails:
- `✅ Email sent successfully!` - Email was sent
- `❌ Failed to send email` - Check the error message
- `⚠️ Resend account limitation` - Need to upgrade or verify domain

**Test Email Delivery:**
```bash
# Use the test endpoint
curl -X POST http://localhost:8080/api/test-email \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

### Gmail SMTP Errors?

**"Username and Password not accepted":**
- Make sure you're using an App Password, not your regular Gmail password
- App Password should be 16 characters without spaces

**"Less secure app access":**
- This is outdated - use App Passwords instead
- App Passwords work with 2FA enabled

### Resend Domain Not Verifying?

**DNS not propagating:**
- Wait 30-60 minutes after adding DNS records
- Use https://dnschecker.org to check if records are visible globally

**Wrong DNS records:**
- Double-check you copied the exact values from Resend
- Make sure there are no extra spaces or characters

---

## 📊 Comparison Table

| Feature | Resend (Free) | Resend (Custom Domain) | Gmail SMTP |
|---------|---------------|------------------------|------------|
| Cost | Free | Free (need domain ~$10/yr) | Free |
| Emails/month | 3,000 | 3,000 | 15,000 |
| Emails/day | 100 | 100 | 500 |
| Send to any email | ❌ | ✅ | ✅ |
| Custom from email | ❌ | ✅ | ❌ |
| Professional | ⚠️ | ✅ | ⚠️ |
| Setup difficulty | Easy | Medium | Easy |
| Best for | Testing | Production | Small apps |

---

## 🚀 Recommended Setup

**For Development/Testing:**
- Use Gmail SMTP (easiest, no domain needed)

**For Production:**
- Use Resend with custom domain (most professional)
- Or upgrade to Resend paid plan if you don't have a domain

---

## 📝 Current Configuration

Your app supports both Resend and Gmail SMTP automatically:
- If `GMAIL_USERNAME` and `GMAIL_APP_PASSWORD` are set → Uses Gmail SMTP
- Otherwise → Uses Resend API
- If neither → Development mode (prints to console)

Check your server startup logs to see which service is active!
