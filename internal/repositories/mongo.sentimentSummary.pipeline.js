[
  {
    $project: {
      _id: 0,
      negative: 1,
      positive: 1,
      category: 1,
      timestamp: 1,
      createdAt: {
        $toDate: {
          $multiply: ["$timestamp", 1000]
        }
      }
    }
  },
  {
    $facet: {
      last15Minutes: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 15 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last30Minutes: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 30 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last1Hour: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 60 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last6Hours: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 6 * 60 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last1Day: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 24 * 60 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last1Week: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 7 * 24 * 60 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ],
      last1Month: [
        {
          $match: {
            "$expr": {
              "$gte": [
                "$createdAt",
                { "$subtract": ["$$NOW", 30 * 24 * 60 * 60 * 1000] }
              ]
            }
          }
        },
        {
          $group: {
            _id: null,
            total_positive: {
              $sum: "$positive"
            },
            total_negative: {
              $sum: "$negative"
            }
          }
        }
      ]
    }
  }
]