# Деплой career_bot на сервер с нуля

## 1. Подключись к серверу
```bash
ssh root@ТВОЙ_IP
```

## 2. Обнови систему и установи нужное
```bash
apt update && apt upgrade -y
apt install -y docker.io docker-compose-v2 nginx certbot python3-certbot-nginx git
```

## 3. Включи Docker чтобы стартовал при загрузке
```bash
systemctl enable docker
systemctl start docker
```
> Это гарантирует что после перезагрузки сервера Docker сам поднимется,
> а `restart: unless-stopped` в docker-compose поднимет контейнер.

## 4. Склонируй проект
```bash
cd /opt
git clone ТВОЙ_РЕПОЗИТОРИЙ career_bot
cd career_bot
```

## 5. Создай .env
```bash
cp .env.example .env
nano .env
```
Заполни:
```
TELEGRAM_BOT_API=123456:ABC-DEF...
WEBHOOK_URL=https://твойдомен.ru
WEBHOOK_PORT=8443
```
Важно:
- `TELEGRAM_BOT_API` указывай без слова `bot`.
- `WEBHOOK_URL` должен быть реальным публичным HTTPS-доменом без пути, например `https://example.ru`.
- Не оставляй значения из примера вроде `your-real-domain.example`.

Быстрая проверка:
```bash
grep -E 'TELEGRAM_BOT_API|WEBHOOK_URL|WEBHOOK_PORT' .env
```

## 6. Настрой Nginx

Скопируй конфиг:
```bash
cp nginx/career_bot.conf /etc/nginx/sites-available/career_bot
```

Замени `YOUR_DOMAIN` на твой домен:
```bash
sed -i 's/YOUR_DOMAIN/твойдомен.ru/g' /etc/nginx/sites-available/career_bot
```

Включи сайт и удали дефолтный:
```bash
ln -s /etc/nginx/sites-available/career_bot /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default
```

## 7. Получи SSL сертификат

Сначала **закомментируй** блок `server { listen 443 ... }` в конфиге (certbot не сможет стартануть без сертификата):
```bash
nano /etc/nginx/sites-available/career_bot
# Оставь только блок listen 80
```

Перезапусти nginx:
```bash
nginx -t && systemctl restart nginx
```

Получи сертификат:
```bash
certbot --nginx -d твойдомен.ru
```

Certbot сам допишет SSL в конфиг. Или если хочешь вручную — раскомментируй блок 443 обратно:
```bash
nano /etc/nginx/sites-available/career_bot
# Раскомментируй блок listen 443
nginx -t && systemctl restart nginx
```

Автопродление сертификата (уже работает по умолчанию, но проверь):
```bash
systemctl enable certbot.timer
```

## 8. Запусти бота
```bash
cd /opt/career_bot
docker compose up -d --build
```

Проверь, что контейнер запущен:
```bash
docker compose ps
docker compose logs --tail=100
```

## 9. Проверь что всё работает
```bash
# Логи бота
docker compose logs -f

# Статус nginx
systemctl status nginx

# Проверь вебхук
curl https://api.telegram.org/bot<ТВОЙ_ТОКЕН>/getWebhookInfo
```

В `getWebhookInfo` поле `url` должно быть вида:
```
https://твойдомен.ru/bot<ТВОЙ_ТОКЕН>
```
Если `url` пустой, содержит тестовый домен или в `last_error_message` есть ошибка Nginx/SSL, исправь `.env`, затем перезапусти:
```bash
docker compose up -d --build
docker compose logs --tail=100
```

## Полезные команды
```bash
# Перезапустить бота
docker compose restart

# Пересобрать после изменений
docker compose up -d --build

# Остановить
docker compose down

# Логи
docker compose logs -f --tail=100
```
