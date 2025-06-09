[
  {
    $match: {
      finsentiment: {
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
        $toDate: "$commit.record.createdat"
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
          $project: {
            corpus: {
              $reduce: {
                input: "$texts",
                initialValue: "",
                in: {
                  $concat: [
                    "$$value",
                    {
                      $cond: [
                        {
                          $eq: ["$$value", ""]
                        },
                        "",
                        " "
                      ]
                    },
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