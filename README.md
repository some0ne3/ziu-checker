## ZIU-checker
Unofficial tool used for fetching latest results from your exams available on ziu.gov.pl website.

### Environment variables
- `ZIU_USERNAME` - your login to ziu.gov.pl
- `ZIU_PASSWORD` - your password to ziu.gov.pl
- `DISCORD_WEBHOOK_URL` - webhook url to discord channel where you want to receive notifications

## Running
(include env variables)
```bash
docker run ghcr.io/some0ne3/ziu-checker:latest
```

## Example discord notification
```
# Egzamin maturalny - XYZ
Placówka: XYZ

## Egzamin: język polski (poz. podstawowy) (Pisemny)
Data wydania dokumentu: 2021-06-08
Numer dokumentu: 123456789
Data egzaminu: 2021-05-04
Kod arkusza: 123456789
Centyle: 78

**Procent: 78**

Punkty: 39.00/50.00
```