echo "Arrêt de tous les conteneurs..."
docker stop $(docker ps -aq)
echo "Suppression de tous les conteneurs..."
docker rm -f $(docker ps -aq)
echo "Suppression de toutes les images..."
docker rmi -f $(docker images -aq)
echo "Suppression de tous les volumes..."
docker volume rm -f $(docker volume ls -q)
echo "Suppression de tous les réseaux non utilisés..."
docker network prune -f
echo "Nettoyage du système Docker..."
docker system prune -af --volumes