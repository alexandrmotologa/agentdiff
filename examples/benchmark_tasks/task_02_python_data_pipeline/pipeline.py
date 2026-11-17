def aggregate_transactions(records):
    total = 0.0
    count = 0
    for record in records:
        total += float(record['amount'])
        count += 1

    avg = total / count if count > 0 else 0.0
    return {
        "total": round(total, 2),
        "count": count,
        "average": round(avg, 2)
    }
