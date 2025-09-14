# Récupérer le token et le stocker sans guillemets
export ACCESS_TOKEN=$(curl -k -X POST http://localhost:8080/realms/demo/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=backend" \
  -d "client_secret=mysecret" \
  -d "username=user" \
  -d "password=test" | jq -r '.access_token')

# Vérifier le token (sans guillemets)
echo $ACCESS_TOKEN

# Utiliser le token pour appeler l'API
curl -X GET -H "Authorization: Bearer $ACCESS_TOKEN" http://localhost:8081/api/public | jq .



curl -X GET -H "Authorization: Bearer $ACCESS_TOKEN" http://localhost:8081/api/opensource | jq .