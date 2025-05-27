export function formatDate(date: string | Date) {
    return dateTimeFormatter(date);
}

export function formatDateTime(date: string | Date) {
    return dateTimeFormatter(date, {
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hour12: false,
    });
}

function dateTimeFormatter(
    date: string | Date,
    customOptions?: Intl.DateTimeFormatOptions,
) {
    const defaultOptions: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
    };

    const options = { ...defaultOptions, ...customOptions };

    const formatter = new Intl.DateTimeFormat('de-DE', options);

    if (typeof date === 'string') {
        date = new Date(date);
    }

    return formatter.format(date);
}
