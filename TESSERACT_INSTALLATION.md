# Tesseract OCR Installation Guide (Windows)

Tesseract is required for passport extraction and CV processing features. This guide covers installation on Windows.

## Quick Install (Recommended)

### Option 1: Windows Installer (Easiest)

1. **Download the installer:**
   - Go to: https://github.com/UB-Mannheim/tesseract/wiki
   - Download: `tesseract-ocr-w64-setup-v5.x.x.exe` (64-bit for Windows)
   - Latest version: **v5.4.x** or higher

2. **Run the installer:**
   - Double-click the `.exe` file
   - Accept the license agreement
   - Choose installation location (default is fine):
     ```
     C:\Program Files (x86)\Tesseract-OCR
     ```
   - Select language data: English (eng) is sufficient
   - Complete the installation

3. **Verify installation:**
   ```powershell
   tesseract --version
   ```
   Should output version info like:
   ```
   tesseract 5.x.x
   leptonica-1.xx.x
   ...
   ```

4. **Configure in .env.local:**
   ```
   TESSERACT_PATH=C:\Program Files (x86)\Tesseract-OCR\tesseract.exe
   ```

   OR leave it empty if Tesseract is in your PATH:
   ```
   TESSERACT_PATH=
   ```

---

### Option 2: Chocolatey Package Manager

If you have Chocolatey installed:

```powershell
# Install Tesseract
choco install tesseract

# Verify
tesseract --version
```

Then set in `.env.local`:
```
TESSERACT_PATH=C:\Program Files\Tesseract-OCR\tesseract.exe
```

---

### Option 3: Docker (Alternative - Not Recommended for Local Dev)

If you prefer containerized approach:

```powershell
docker run -it -v C:\your\documents:/documents ubuntu:22.04 bash
apt-get update
apt-get install -y tesseract-ocr
tesseract /documents/image.jpg stdout
```

---

## Verify Installation

### Command Line Test
```powershell
# Test Tesseract is accessible
tesseract --version

# Test OCR on a sample image (if you have one)
tesseract "C:\path\to\image.jpg" output.txt
```

### Test within Application
```powershell
# Start the API
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
make run-api
```

Then try the OCR endpoint:
```powershell
$headers = @{"Authorization" = "Bearer your-jwt-token"}
$file = Get-Item "C:\path\to\passport.jpg"
$form = @{file = $file}
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/candidates/passport/parse-preview" `
  -Method Post -Headers $headers -Form $form
```

---

## Troubleshooting

### "tesseract: The term 'tesseract' is not recognized"

**Problem:** Tesseract is not in your PATH

**Solution:**

#### Method 1: Add to PATH Manually
1. Open System Properties:
   - Windows + X → System
   - Advanced system settings
   - Environment Variables

2. Add Tesseract to PATH:
   - Click "New" under "User variables"
   - Variable name: `PATH`
   - Variable value: `C:\Program Files (x86)\Tesseract-OCR`

3. Restart PowerShell/terminal

#### Method 2: Set TESSERACT_PATH in .env.local
```
TESSERACT_PATH=C:\Program Files (x86)\Tesseract-OCR\tesseract.exe
```

---

### "Tesseract is unavailable" in API

**Problem:** Application can't find Tesseract

**Solutions:**
1. Verify installation: `tesseract --version`
2. Check .env.local has correct `TESSERACT_PATH`
3. Restart API: `Ctrl+C` then `make run-api`
4. Check API logs for exact error message

---

### OCR Extraction Returns Empty Results

**Problem:** Tesseract runs but extracts nothing

**Likely Causes:**
- Image quality is too poor (blurry, low contrast)
- Image is not a valid passport
- Image is upside down or rotated

**Solutions:**
- Use high-quality images (300+ DPI)
- Ensure images are well-lit and clear
- Try rotating the image 90°
- Check Tesseract can read the image:
  ```powershell
  tesseract "C:\path\to\image.jpg" output.txt
  notepad output.txt
  ```

---

## Language Support

By default, only English (eng) is installed. To add more languages:

### Install Additional Languages

1. Download language files from:
   - https://github.com/UB-Mannheim/tesseract/wiki#tesseract-stable-releases

2. Copy to Tesseract folder:
   ```powershell
   Copy-Item "*.traineddata" "C:\Program Files (x86)\Tesseract-OCR\tessdata\"
   ```

3. Set in .env.local:
   ```
   OCR_LANGUAGE=eng+fra+deu
   ```
   (English + French + German)

---

## Performance Optimization

### First Run vs Subsequent Runs

- **First extraction:** ~3-5 seconds (Tesseract initialization)
- **Subsequent extractions:** ~1-2 seconds (cached)
- **Cache TTL:** 15 minutes

### Tips for Better Performance

1. **Use high-quality images:**
   - Resolution: 300+ DPI
   - Format: JPG (85% quality) or PNG
   - Size: Under 5MB is optimal

2. **Batch operations:**
   - If processing multiple passports, do one at a time
   - Let cache fill before heavy usage

3. **Monitor resources:**
   - Tesseract can use significant CPU during OCR
   - Consider disabling other apps during heavy OCR use

---

## System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| OS | Windows 7 SP1 | Windows 10 / 11 |
| Disk Space | 100MB | 500MB |
| RAM | 2GB | 4GB+ |
| CPU | 2-core | 4-core+ |
| Tesseract Version | 4.1 | 5.0+ |

---

## Uninstall

### Via Installer
1. Go to Control Panel → Programs → Programs and Features
2. Find "Tesseract OCR"
3. Click Uninstall

### Via Chocolatey
```powershell
choco uninstall tesseract
```

### Manual
```powershell
Remove-Item "C:\Program Files (x86)\Tesseract-OCR" -Recurse
```

---

## Support & Resources

- **Official Repository:** https://github.com/UB-Mannheim/tesseract/wiki
- **Tesseract Project:** https://github.com/tesseract-ocr/tesseract
- **Language Packs:** https://github.com/tesseract-ocr/tessdata
- **Documentation:** https://tesseract-ocr.github.io/

---

## Next Steps

Once Tesseract is installed:

1. Verify with: `tesseract --version`
2. Update `.env.local` if needed
3. Restart the API: `make run-api`
4. Test OCR endpoint with a passport image
5. Check `CREDENTIALS_AND_ENDPOINTS.md` for API examples

**You're all set! 🎉**
