#!/usr/bin/env python3

import logging
import random
from functools import wraps
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

logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO)
logger = logging.getLogger(__name__)
logging.getLogger("httpx").setLevel(logging.WARNING)

messages_path = Path(__file__).parent / "messages.yaml"
with open(messages_path, encoding="utf-8") as f:
    MESSAGES = yaml.safe_load(f)

db = Database(DB_PATH)

# Users awaiting sticker recording (in-memory, dev tool)
awaiting_sticker: set[int] = set()


def admin_only(func):
    """Decorator to restrict command to admins only."""

    @wraps(func)
    async def wrapper(update: Update, context: ContextTypes.DEFAULT_TYPE):
        if not update.effective_user or not await db.is_admin(update.effective_user.id):
            return
        return await func(update, context)

    return wrapper

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


def pick_random_time(time_intersection: str) -> str:
    """Pick random time from time intersection binary string.

    Example: "101001" -> picks random time with 5min interval
    in first hour of randomly selected range where bit is 1.
    """
    indices = [i for i, bit in enumerate(time_intersection) if bit == "1"]

    index = random.choice(indices)
    time_range = TIME_RANGES[index]

    begin = [x.split(':')[0] for x in time_range.split(' -- ')][0]
    minutes = random.choice(list(range(12))) * 5

    return f'{begin}:{minutes:02d}'


def create_time_keyboard(selected_times: set[str]) -> InlineKeyboardMarkup:
    """Create keyboard with time ranges and confirm button (if any selected)."""
    keyboard = []
    for i in range(0, len(TIME_RANGES), 2):
        row = []
        for time_range in TIME_RANGES[i : i + 2]:
            text = f"> {time_range} <" if time_range in selected_times else time_range
            row.append(InlineKeyboardButton(text, callback_data=f"time_{time_range}"))
        keyboard.append(row)

    if selected_times:
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

    await db.save_user(
        telegram_id=user.id,
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
        is_bot=user.is_bot,
        language_code=user.language_code,
        is_premium=user.is_premium,
    )

    await db.set_user_state(user.id, "awaiting_sex")

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

    await db.set_user_sex(user.id, sex)
    await db.set_user_state(user.id, "awaiting_about")

    await query.edit_message_text(MESSAGES["sex_selected"])


async def text_message_handler(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle text messages based on user state."""
    if not update.message or not update.effective_user or not update.message.text:
        return

    user = update.effective_user
    state = await db.get_user_state(user.id)

    if state == "awaiting_about":
        await db.set_user_about(user.id, update.message.text)
        await db.set_user_state(user.id, "awaiting_time")

        binary_str = await db.get_time_ranges(user.id)
        selected_times = binary_to_set(binary_str)
        keyboard = create_time_keyboard(selected_times)

        await update.message.reply_text(MESSAGES["about_received"]["message"], reply_markup=keyboard)


async def time_button_callback(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle time range button clicks."""
    query = update.callback_query
    if not query or not query.data or not query.from_user:
        return

    await query.answer()

    user = query.from_user

    if query.data == "confirm_time":
        await db.set_user_state(user.id, "completed")
        await query.edit_message_text(MESSAGES["completed"]["message"])
        return

    time_range = query.data.replace("time_", "")

    binary_str = await db.get_time_ranges(user.id)
    selected = binary_to_set(binary_str)

    if time_range in selected:
        selected.remove(time_range)
    else:
        selected.add(time_range)

    new_binary = set_to_binary(selected)
    await db.save_time_ranges(user.id, new_binary)

    keyboard = create_time_keyboard(selected)

    await query.edit_message_reply_markup(reply_markup=keyboard)


async def confirm_meeting_callback(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle meeting confirmation button clicks."""
    query = update.callback_query
    if not query or not query.data or not query.from_user or not context.bot:
        return

    await query.answer()

    meeting_id = int(query.data.replace("confirm_meeting_", ""))
    telegram_id = query.from_user.id

    if await db.confirm_meeting(meeting_id, telegram_id):
        cancel_keyboard = InlineKeyboardMarkup([
            [InlineKeyboardButton(
                MESSAGES["meet"]["cancel_button"],
                callback_data=f"cancel_meeting_{meeting_id}"
            )]
        ])

        original_text = query.message.text
        new_text = f"{original_text}\n\n{MESSAGES['meet']['confirmed']}"
        await query.edit_message_text(
            text=new_text,
            reply_markup=cancel_keyboard
        )

        partner_id = await db.get_partner_telegram_id(meeting_id, telegram_id)
        if partner_id:
            try:
                await context.bot.send_message(
                    chat_id=partner_id,
                    text=MESSAGES["meet"]["partner_confirmed"]
                )
            except Exception as e:
                logger.error(f"Failed to send confirmation to partner {partner_id}: {e}")

        if await db.both_confirmed(meeting_id):
            meeting = await db.get_meeting_by_id(meeting_id)
            if meeting:
                places = await db.get_all_places()
                place_desc = next((p.description for p in places if p.id == meeting.place_id), "")

                final_message = MESSAGES["meet"]["both_confirmed"].format(
                    place=place_desc,
                    time=meeting.time
                )

                cancel_keyboard_final = InlineKeyboardMarkup([
                    [InlineKeyboardButton(
                        MESSAGES["meet"]["cancel_button"],
                        callback_data=f"cancel_meeting_{meeting_id}"
                    )]
                ])

                try:
                    await context.bot.send_message(
                        chat_id=telegram_id,
                        text=final_message,
                        reply_markup=cancel_keyboard_final
                    )
                except Exception as e:
                    logger.error(f"Failed to send final confirmation to {telegram_id}: {e}")

                if partner_id:
                    try:
                        await context.bot.send_message(
                            chat_id=partner_id,
                            text=final_message,
                            reply_markup=cancel_keyboard_final
                        )
                    except Exception as e:
                        logger.error(f"Failed to send final confirmation to {partner_id}: {e}")


async def cancel_meeting_callback(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle meeting cancellation button clicks."""
    query = update.callback_query
    if not query or not query.data or not query.from_user or not context.bot:
        return

    await query.answer()

    meeting_id = int(query.data.replace("cancel_meeting_", ""))
    telegram_id = query.from_user.id

    if await db.cancel_meeting(meeting_id, telegram_id):
        partner_username = await db.get_partner_username(meeting_id, telegram_id)
        user_username = await db.get_user_username_by_telegram_id(telegram_id)

        await query.edit_message_text(
            text=MESSAGES["meet"]["cancelled"].format(
                partner_username=partner_username or "unknown"
            )
        )

        partner_id = await db.get_partner_telegram_id(meeting_id, telegram_id)
        if partner_id:
            try:
                await context.bot.send_message(
                    chat_id=partner_id,
                    text=MESSAGES["meet"]["partner_cancelled"].format(
                        partner_username=user_username or "unknown"
                    )
                )
            except Exception as e:
                logger.error(f"Failed to send cancellation to partner {partner_id}: {e}")


@admin_only
async def stickerid_command(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /stickerid command - wait for sticker and print its file_id."""
    if not update.message or not update.effective_user:
        return

    awaiting_sticker.add(update.effective_user.id)
    await update.message.reply_text(MESSAGES["stickerid"]["prompt"])


@admin_only
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
    await update.message.reply_text(f"```\n{sticker.file_id}\n```", parse_mode="Markdown")


@admin_only
async def promote_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /promote command - grant admin rights to a user."""
    if not update.message:
        return

    if not context.args:
        await update.message.reply_text(MESSAGES["promote"]["usage"])
        return

    username = context.args[0].lstrip("@")
    target_user = await db.get_user_by_username(username)

    if not target_user:
        await update.message.reply_text(MESSAGES["promote"]["user_not_found"].format(username=username))
        return

    if target_user.is_admin:
        await update.message.reply_text(MESSAGES["promote"]["already_admin"].format(username=username))
        return

    await db.set_admin(target_user.telegram_id, True)
    await update.message.reply_text(MESSAGES["promote"]["success"].format(username=username))


@admin_only
async def demote_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /demote command - revoke admin rights from a user."""
    if not update.message:
        return

    if not context.args:
        await update.message.reply_text(MESSAGES["demote"]["usage"])
        return

    username = context.args[0].lstrip("@")
    target_user = await db.get_user_by_username(username)

    if not target_user:
        await update.message.reply_text(MESSAGES["demote"]["user_not_found"].format(username=username))
        return

    if not target_user.is_admin:
        await update.message.reply_text(MESSAGES["demote"]["not_admin"].format(username=username))
        return

    await db.set_admin(target_user.telegram_id, False)
    await update.message.reply_text(MESSAGES["demote"]["success"].format(username=username))


@admin_only
async def match_command(update: Update, _: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /match command - match all verified users into pairs."""
    if not update.message:
        return

    users = await db.get_verified_users()

    if len(users) < 2:
        await update.message.reply_text(MESSAGES["match"]["not_enough_users"])
        return

    sticker_msg = await update.message.reply_sticker(
        "CAACAgIAAxkBAANtaYKDDtR5d1478iPkCrZr2xnZOpMAAgIBAAJWnb0KTuJsgctA5P84BA"
    )

    pairs, full_matches = match_people(users)

    await db.clear_pairs()

    for i, j, score, time_intersection in pairs:
        dill = users[i]
        doe = users[j]

        await db.save_pair(dill.id, doe.id, score, time_intersection, is_fullmatch=False)

    for i, j, score in full_matches:
        dill = users[i]
        doe = users[j]

        await db.save_pair(dill.id, doe.id, score, "000000", is_fullmatch=True)

    await sticker_msg.delete()

    full_info = f"\n\nполных совпадений (без общего времени): {len(full_matches)}" if full_matches else ""
    await update.message.reply_text(
        MESSAGES["match"]["success"].format(pairs=len(pairs), users=len(users), full_info=full_info)
    )


@admin_only
async def meet_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /meet command - send meeting notifications to all pairs."""
    if not update.message or not context.bot:
        return

    regular_pairs = await db.get_regular_pairs()
    fullmatch_pairs = await db.get_full_pairs()

    if not regular_pairs and not fullmatch_pairs:
        await update.message.reply_text(MESSAGES["meet"]["no_pairs"])
        return

    places = await db.get_all_places()
    if not places and regular_pairs:
        await update.message.reply_text(MESSAGES["meet"]["no_places"])
        return

    count = 0

    for pair in regular_pairs:
        place = random.choice(places)
        meeting_time = pick_random_time(pair.time_intersection)

        meeting_id = await db.save_meeting(pair.id, place.id, meeting_time)

        dill = await db.get_user_by_id(pair.dill_id)
        doe = await db.get_user_by_id(pair.doe_id)

        if dill and doe:
            message = MESSAGES["meet"]["notification"].format(
                place=place.description,
                time=meeting_time
            )

            keyboard = InlineKeyboardMarkup([
                [
                    InlineKeyboardButton(
                        MESSAGES["meet"]["confirm_button"],
                        callback_data=f"confirm_meeting_{meeting_id}"
                    ),
                    InlineKeyboardButton(
                        MESSAGES["meet"]["cancel_button"],
                        callback_data=f"cancel_meeting_{meeting_id}"
                    )
                ]
            ])

            try:
                await context.bot.send_message(
                    chat_id=dill.telegram_id,
                    text=message,
                    reply_markup=keyboard
                )
            except Exception as e:
                logger.error(f"Failed to send message to {dill.telegram_id}: {e}")

            try:
                await context.bot.send_message(
                    chat_id=doe.telegram_id,
                    text=message,
                    reply_markup=keyboard
                )
            except Exception as e:
                logger.error(f"Failed to send message to {doe.telegram_id}: {e}")

            count += 1

    for pair in fullmatch_pairs:
        dill = await db.get_user_by_id(pair.dill_id)
        doe = await db.get_user_by_id(pair.doe_id)

        if dill and doe:
            try:
                await context.bot.send_message(
                    chat_id=dill.telegram_id,
                    text=MESSAGES["meet"]["full_match"].format(
                        partner_username=doe.username or "unknown"
                    )
                )
            except Exception as e:
                logger.error(f"Failed to send full match to {dill.telegram_id}: {e}")

            try:
                await context.bot.send_message(
                    chat_id=doe.telegram_id,
                    text=MESSAGES["meet"]["full_match"].format(
                        partner_username=dill.username or "unknown"
                    )
                )
            except Exception as e:
                logger.error(f"Failed to send full match to {doe.telegram_id}: {e}")

            count += 1

    await update.message.reply_text(MESSAGES["meet"]["success"].format(count=count))


async def post_init(application: Application) -> None:
    """Initialize database after application starts."""
    await db.init_db()
    logger.info("Database initialized")


def main() -> None:
    application = Application.builder().token(TELEGRAM_TOKEN).post_init(post_init).build()

    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("match", match_command))
    application.add_handler(CommandHandler("meet", meet_command))
    application.add_handler(CommandHandler("stickerid", stickerid_command))
    application.add_handler(CommandHandler("promote", promote_command))
    application.add_handler(CommandHandler("demote", demote_command))

    application.add_handler(CallbackQueryHandler(sex_callback, pattern="^sex_"))
    application.add_handler(CallbackQueryHandler(time_button_callback, pattern="^(time_|confirm_time)"))
    application.add_handler(CallbackQueryHandler(confirm_meeting_callback, pattern="^confirm_meeting_"))
    application.add_handler(CallbackQueryHandler(cancel_meeting_callback, pattern="^cancel_meeting_"))

    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, text_message_handler))
    application.add_handler(MessageHandler(filters.Sticker.ALL, sticker_handler))

    logger.info("Starting bot...")
    application.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    main()
