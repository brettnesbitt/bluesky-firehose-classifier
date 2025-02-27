PROFILE="default"

eval "$(dotenvx run -f .env --quiet -- printenv)"

if [[ -n "${TEXT_CATEGORY_CLASSIFIER}" ]]; then \
	PROFILE="${PROFILE},text-category-classifier"; \
fi; \
if [[ -n "${TEXT_FIN_SENTIMENT_CLASSIFIER}" ]]; then \
    PROFILE="${PROFILE},text-fin-sentiment-classifier"; \
fi;

echo ${PROFILE}