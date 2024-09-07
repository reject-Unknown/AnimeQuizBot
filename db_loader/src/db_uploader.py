from typing import Dict, Any, List
from jikanpy import Jikan
from pymongo.collection import Collection
import tqdm
import logging


from pymongo import MongoClient
from ratelimit import limits, sleep_and_retry


JIKAN = Jikan()
LOG = logging.getLogger()
LEVELS:  int = 3
MAX_ON_PAGE: int = 25


class JikanLimiterAdapter:
    pass


@sleep_and_retry
@limits(calls=1, period=1)
def limiter_jikan():
    return JIKAN


def _extract_title(anime: Dict[str, Any]):
    titles: Dict[str, str] = {title["type"]: title["title"] for title in anime["titles"]}
    return titles.get("English", titles["Default"])

def _get_anime_on_level(level: int, level_gap: int):
    limit: int = min(MAX_ON_PAGE, level_gap)
    skip_pages: int = ((level - 1) * level_gap) // limit
    start_from_first_page: int = (level - 1) * level_gap - skip_pages * limit
    start_page: int = skip_pages + 1
    if start_from_first_page != 0:
        yield from limiter_jikan().top(type="anime", parameters={"limit": limit, "page": start_page, "filter": "bypopularity"})["data"][start_from_first_page:]
        start_page += 1

    anime_without_first_page: int = level_gap - start_from_first_page
    full_pages, last_page_size = divmod(anime_without_first_page, limit)
    for page in range(start_page, start_page + full_pages):
        yield from limiter_jikan().top(type="anime", parameters={"limit": limit, "page": page, "filter": "bypopularity"})["data"]

    if last_page_size != 0:
        yield from limiter_jikan().top(type="anime", parameters={"limit": limit, "page": last_page_size, "filter": "bypopularity"})["data"]


def get_all_main_characters(cluster: MongoClient, level_gap: int):
    database = cluster["QuizDB"]

    LOG.info("Init completed")
    characters: Dict[int, Dict[str, Any]] = {}
    for level in range(1, LEVELS + 1):
        LOG.info(f"Start processing {level}th level")
        collection: Collection = database[f"CharactersLevel{level}"]
        collection.delete_many({})
        for anime in tqdm.tqdm(_get_anime_on_level(level, level_gap), desc=f"Level {level}", total=level_gap):
            anime_characters: List[Any[str, Any]] = limiter_jikan().anime(
                anime["mal_id"], extension="characters")["data"]
            
            anime_name: str = _extract_title(anime)
            for character in anime_characters:
                if character["role"] != "Main" or character["character"]["mal_id"] in characters:
                    continue

                result_character: Dict[str, Any] = {
                    "id": character["character"]["mal_id"],
                    "name": character["character"]["name"],
                    "image_url": character["character"]["images"]["jpg"]["image_url"],
                    "anime_url": anime["url"],
                    "anime_name": anime_name
                }
                characters[character["character"]["mal_id"]] = result_character

        collection.insert_many(characters.values())


