[
  {
    $match: {
      fin_sentiment: {
        $in: ["positive", "negative"]
      },
      categories: {
        $in: [
          "politics",
          "war",
          "education",
          "government"
        ]
      }
    }
  },
  {
    $project: {
      _id: 0,
      text: "$commit.record.text",
      createdAtDate: {
        $toDate: "$commit.record.created_at"
      }
    }
  },
  {
    $sort: {
      createdAtDate: 1
    }
  },
  {
    $facet: {
      last15Minutes: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAtDate",
                { "$subtract": ["$$NOW", 15 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            texts: {
              $push: "$text"
            }
          }
        },
        {
          "$project": {
            "corpus": {
              "$reduce": {
                "input": "$texts",
                "initialValue": "",
                "in": {
                  "$concat": [
                    "$$value",
                    // Add a space separator, but only if the string isn't empty
                    {
                      "$cond": {
                        "if": { "$eq": [ "$$value", "" ] },
                        "then": "",
                        "else": " "
                      }
                    },
                    // Add the prefix for every item
                    "[POST]: ",
                    "$$this"
                  ]
                }
              }
            }
          }
        }
      ],
      last30Minutes: [
        {
          $match: {
            createdAtDate: {
              $gte: "$$NOW"
            }
          }
        },
        {
          $group: {
            _id: null,
            texts: {
              $push: "$text"
            }
          }
        },
        {
          $project: {
            corpus: ""
          }
        }
      ],
      last1Hour: [
        {
          $match: {
            createdAtDate: {
              $gte: "$$NOW"
            }
          }
        },
        {
          $group: {
            _id: null,
            texts: {
              $push: "$text"
            }
          }
        },
        {
          $project: {
            corpus: ""
          }
        }
      ],
      last1Day: [
        {
          $match: {
            createdAtDate: {
              $gte: "$$NOW"
            }
          }
        },
        {
          $group: {
            _id: null,
            texts: {
              $push: "$text"
            }
          }
        },
        {
          $project: {
            corpus: ""
          }
        }
      ]
    }
  }
]