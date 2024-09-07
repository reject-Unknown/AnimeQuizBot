import db_uploader
import utils

from pymongo import MongoClient

CONNECTION_STR: str = "mongodb://{user}:{password}@127.0.0.1:22222/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.3.0"

if __name__ == "__main__":
    args = utils.parse_args()

    cluster: MongoClient = MongoClient(CONNECTION_STR.format(user=args.user, password=args.password))
    db_uploader.get_all_main_characters(cluster, 50)
