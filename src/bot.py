#!/usr/bin/env python3

import logging
from pathlib import Path

import yaml
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update
from telegram.ext import (
    Application,
    CallbackQueryHandler,
    CommandHandler,
    ContextTypes,
    MessageHandler,
    filters,
)

from config import DB_PATH, TELEGRAM_TOKEN
from database import Database
from matcher import match_people

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    level=logging.INFO,
)
logger = logging.getLogger(__name__)
logging.getLogger("httpx").setLevel(logging.WARNING)

messages_path = Path(__file__).parent / "messages.yaml"
with open(messages_path, encoding="utf-8") as f:
    MESSAGES = yaml.safe_load(f)

db = Database(DB_PATH)

# Users awaiting sticker recording (in-memory, dev tool)
awaiting_sticker: set[int] = set()

TIME_RANGES = [
    "10:00 -- 12:00",
    "12:00 -- 14:00",
    "14:00 -- 16:00",
    "16:00 -- 18:00",
    "18:00 -- 20:00",
    "20:00 -- 22:00",
]


def binary_to_set(binary_str: str) -> set[str]:
    """Convert binary string to set of selected time ranges."""
    selected = set()
    for i, bit in enumerate(binary_str):
        if bit == "1" and i < len(TIME_RANGES):
            selected.add(TIME_RANGES[i])
    return selected


def set_to_binary(selected: set[str]) -> str:
    """Convert set of selected time ranges to binary string."""
    binary = []
    for time_range in TIME_RANGES:
        binary.append("1" if time_range in selected else "0")
    return "".join(binary)


def create_time_keyboard(
    selected_times: set[str], include_confirm: bool = True
) -> InlineKeyboardMarkup:
    """Create keyboard with time ranges and optional confirm button."""
    keyboard = []
    for i in range(0, len(TIME_RANGES), 2):
        row = []
        for time_range in TIME_RANGES[i : i + 2]:
            text = f"> {time_range} <" if time_range in selected_times else time_range
            row.append(InlineKeyboardButton(text, callback_data=f"time_{time_range}"))
        keyboard.append(row)

    if include_confirm:
        keyboard.append(
            [
                InlineKeyboardButton(
                    MESSAGES["buttons"]["confirm"], callback_data="confirm_time"
                )
            ]
        )

    return InlineKeyboardMarkup(keyboard)


async def start(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /start command - welcome and gender selection."""
    if not update.message or not update.effective_user:
        return

    user = update.effective_user

    db.save_user(
        telegram_id=user.id,
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
        is_bot=user.is_bot,
        language_code=user.language_code,
        is_premium=user.is_premium,
    )

    db.set_user_state(user.id, "awaiting_sex")

    await update.message.reply_text(MESSAGES["start"]["welcome"])

    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    MESSAGES["buttons"]["sex"]["male"], callback_data="sex_male"
                ),
                InlineKeyboardButton(
                    MESSAGES["buttons"]["sex"]["female"], callback_data="sex_female"
                ),
            ]
        ]
    )

    await update.message.reply_text(MESSAGES["start"]["ask_sex"], reply_markup=keyboard)


async def sex_callback(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle gender selection."""
    query = update.callback_query
    if not query or not query.data or not query.from_user:
        return

    await query.answer()

    user = query.from_user
    sex = "male" if query.data == "sex_male" else "female"

    db.set_user_sex(user.id, sex)
    db.set_user_state(user.id, "awaiting_about")

    await query.edit_message_text(MESSAGES["sex_selected"])


async def text_message_handler(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle text messages based on user state."""
    if not update.message or not update.effective_user or not update.message.text:
        return

    user = update.effective_user
    state = db.get_user_state(user.id)

    if state == "awaiting_about":
        db.set_user_about(user.id, update.message.text)
        db.set_user_state(user.id, "awaiting_time")

        binary_str = db.get_time_ranges(user.id)
        selected_times = binary_to_set(binary_str)
        keyboard = create_time_keyboard(selected_times)

        await update.message.reply_text(
            MESSAGES["about_received"]["message"],
            reply_markup=keyboard,
        )


async def time_button_callback(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle time range button clicks."""
    query = update.callback_query
    if not query or not query.data or not query.from_user:
        return

    await query.answer()

    user = query.from_user

    if query.data == "confirm_time":
        db.set_user_state(user.id, "completed")
        await query.edit_message_text(MESSAGES["completed"]["message"])
        return

    time_range = query.data.replace("time_", "")

    binary_str = db.get_time_ranges(user.id)
    selected = binary_to_set(binary_str)

    if time_range in selected:
        selected.remove(time_range)
    else:
        selected.add(time_range)

    new_binary = set_to_binary(selected)
    db.save_time_ranges(user.id, new_binary)

    keyboard = create_time_keyboard(selected)

    await query.edit_message_reply_markup(reply_markup=keyboard)


async def stickerid_command(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /stickerid command - wait for sticker and print its file_id."""
    if not update.message or not update.effective_user:
        return

    awaiting_sticker.add(update.effective_user.id)
    await update.message.reply_text("Send me a sticker and I'll print its file_id.")


async def sticker_handler(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle sticker messages."""
    if not update.message or not update.effective_user or not update.message.sticker:
        return

    user_id = update.effective_user.id
    if user_id not in awaiting_sticker:
        return

    awaiting_sticker.discard(user_id)
    sticker = update.message.sticker
    logger.info(f"Sticker file_id: {sticker.file_id}")
    await update.message.reply_text(
        f"```\n{sticker.file_id}\n```", parse_mode="Markdown"
    )


async def match_command(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /match command - match all verified users into pairs."""
    if not update.message:
        return

    users = db.get_verified_users()

    if len(users) < 2:
        await update.message.reply_text("Not enough verified users to match.")
        return

    sticker_msg = await update.message.reply_sticker(
        "CAACAgIAAxkBAANtaYKDDtR5d1478iPkCrZr2xnZOpMAAgIBAAJWnb0KTuJsgctA5P84BA"
    )

    pairs = match_people(users)

    db.clear_pairs()

    for i, j, score, time_intersection in pairs:
        dill = users[i]
        doe = users[j]

        db.save_pair(dill.id, doe.id, score, time_intersection)

    await sticker_msg.delete()
    await update.message.reply_text(
        f"Matching complete! Created {len(pairs)} pairs from {len(users)} users."
    )


def main() -> None:
    application = Application.builder().token(TELEGRAM_TOKEN).build()

    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("match", match_command))
    application.add_handler(CommandHandler("stickerid", stickerid_command))

    application.add_handler(CallbackQueryHandler(sex_callback, pattern="^sex_"))
    application.add_handler(
        CallbackQueryHandler(time_button_callback, pattern="^(time_|confirm_time)")
    )

    application.add_handler(
        MessageHandler(filters.TEXT & ~filters.COMMAND, text_message_handler)
    )
    application.add_handler(MessageHandler(filters.Sticker.ALL, sticker_handler))

    logger.info("Starting bot...")
    application.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    main()
