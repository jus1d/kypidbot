import sqlite3
from dataclasses import dataclass
from typing import Optional


@dataclass
class User:
    id: int
    telegram_id: int
    username: Optional[str]
    first_name: Optional[str]
    last_name: Optional[str]
    is_bot: bool
    language_code: Optional[str]
    is_premium: bool
    sex: Optional[str]
    about: str
    state: str
    time_ranges: str
    is_admin: bool = False


class Database:
    def __init__(self, db_path: str = "kypidbot.db"):
        self.db_path = db_path
        self.init_db()

    def get_connection(self) -> sqlite3.Connection:
        """Get database connection."""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

    def init_db(self) -> None:
        """Initialize database schema."""
        with self.get_connection() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS users (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    telegram_id INTEGER UNIQUE NOT NULL,
                    username TEXT,
                    first_name TEXT,
                    last_name TEXT,
                    is_bot INTEGER NOT NULL DEFAULT 0,
                    language_code TEXT,
                    is_premium INTEGER DEFAULT false,
                    sex TEXT,
                    about TEXT DEFAULT '',
                    state TEXT NOT NULL DEFAULT 'start',
                    time_ranges TEXT NOT NULL DEFAULT '000000',
                    is_admin INTEGER NOT NULL DEFAULT 0
                )
            """)
            # Migration: add is_admin column to existing databases
            cursor = conn.execute("PRAGMA table_info(users)")
            columns = [row[1] for row in cursor.fetchall()]
            if "is_admin" not in columns:
                conn.execute(
                    "ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0"
                )
            conn.execute("""
                CREATE TABLE IF NOT EXISTS pairs (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    dill_id INTEGER NOT NULL,
                    doe_id INTEGER NOT NULL,
                    score REAL NOT NULL,
                    time_intersection TEXT NOT NULL,
                    FOREIGN KEY (dill_id) REFERENCES users(id),
                    FOREIGN KEY (doe_id) REFERENCES users(id)
                )
            """)
            conn.execute("""
                CREATE TABLE IF NOT EXISTS places (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    description TEXT NOT NULL
                )
            """)
            conn.execute("""
                CREATE TABLE IF NOT EXISTS meetings (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    pair_id INTEGER NOT NULL,
                    place_id INTEGER NOT NULL,
                    time TEXT NOT NULL,
                    dill_confirmed INTEGER DEFAULT 0,
                    doe_confirmed INTEGER DEFAULT 0,
                    dill_cancelled INTEGER DEFAULT 0,
                    doe_cancelled INTEGER DEFAULT 0,
                    FOREIGN KEY (pair_id) REFERENCES pairs(id),
                    FOREIGN KEY (place_id) REFERENCES places(id)
                )
            """)
            conn.commit()

    def save_user(
        self,
        telegram_id: int,
        username: Optional[str],
        first_name: Optional[str],
        last_name: Optional[str],
        is_bot: bool,
        language_code: Optional[str],
        is_premium: Optional[bool],
    ) -> None:
        """Save or update user information."""
        with self.get_connection() as conn:
            conn.execute(
                """
                INSERT INTO users (
                    telegram_id, username, first_name, last_name,
                    is_bot, language_code, is_premium
                ) VALUES (?, ?, ?, ?, ?, ?, ?)
                ON CONFLICT(telegram_id) DO UPDATE SET
                    username = excluded.username,
                    first_name = excluded.first_name,
                    last_name = excluded.last_name,
                    is_bot = excluded.is_bot,
                    language_code = excluded.language_code,
                    is_premium = excluded.is_premium
                """,
                (
                    telegram_id,
                    username,
                    first_name,
                    last_name,
                    1 if is_bot else 0,
                    language_code,
                    1 if is_premium is True else (0 if is_premium is False else None),
                ),
            )
            conn.commit()

    def get_time_ranges(self, telegram_id: int) -> str:
        """Get user's time ranges as binary string (e.g., '101010')."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT time_ranges FROM users WHERE telegram_id = ?",
                (telegram_id,),
            )
            row = cursor.fetchone()
            return row["time_ranges"] if row else "000000"

    def save_time_ranges(self, telegram_id: int, time_ranges: str) -> None:
        """Save user's time ranges as binary string."""
        with self.get_connection() as conn:
            conn.execute(
                "UPDATE users SET time_ranges = ? WHERE telegram_id = ?",
                (time_ranges, telegram_id),
            )
            conn.commit()

    def get_user(self, telegram_id: int) -> Optional[sqlite3.Row]:
        """Get user by telegram_id."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT * FROM users WHERE telegram_id = ?",
                (telegram_id,),
            )
            return cursor.fetchone()

    def get_user_by_username(self, username: str) -> Optional[sqlite3.Row]:
        """Get user by username."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT * FROM users WHERE username = ?",
                (username,),
            )
            return cursor.fetchone()

    def set_user_sex(self, telegram_id: int, sex: str) -> None:
        """Set user's sex (male/female)."""
        with self.get_connection() as conn:
            conn.execute(
                "UPDATE users SET sex = ? WHERE telegram_id = ?",
                (sex, telegram_id),
            )
            conn.commit()

    def set_user_about(self, telegram_id: int, about: str) -> None:
        """Set user's about/introduction text."""
        with self.get_connection() as conn:
            conn.execute(
                "UPDATE users SET about = ? WHERE telegram_id = ?",
                (about, telegram_id),
            )
            conn.commit()

    def set_user_state(self, telegram_id: int, state: str) -> None:
        """Set user's current state in the flow."""
        with self.get_connection() as conn:
            conn.execute(
                "UPDATE users SET state = ? WHERE telegram_id = ?",
                (state, telegram_id),
            )
            conn.commit()

    def get_user_state(self, telegram_id: int) -> str:
        """Get user's current state."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT state FROM users WHERE telegram_id = ?",
                (telegram_id,),
            )
            row = cursor.fetchone()
            return row["state"] if row else "start"

    def is_admin(self, telegram_id: int) -> bool:
        """Check if user is an admin."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT is_admin FROM users WHERE telegram_id = ?",
                (telegram_id,),
            )
            row = cursor.fetchone()
            return bool(row["is_admin"]) if row else False

    def set_admin(self, telegram_id: int, is_admin: bool) -> None:
        """Set user's admin status."""
        with self.get_connection() as conn:
            conn.execute(
                "UPDATE users SET is_admin = ? WHERE telegram_id = ?",
                (1 if is_admin else 0, telegram_id),
            )
            conn.commit()

    def get_verified_users(self) -> list[User]:
        """Get all users who completed verification."""
        with self.get_connection() as conn:
            cursor = conn.execute("SELECT * FROM users WHERE state = 'completed'")
            return [
                User(
                    id=row["id"],
                    telegram_id=row["telegram_id"],
                    username=row["username"],
                    first_name=row["first_name"],
                    last_name=row["last_name"],
                    is_bot=bool(row["is_bot"]),
                    language_code=row["language_code"],
                    is_premium=bool(row["is_premium"]) if row["is_premium"] is not None else False,
                    sex=row["sex"],
                    about=row["about"],
                    state=row["state"],
                    time_ranges=row["time_ranges"],
                    is_admin=bool(row["is_admin"]),
                )
                for row in cursor.fetchall()
            ]

    def clear_pairs(self) -> None:
        """Delete all existing pairs."""
        with self.get_connection() as conn:
            conn.execute("DELETE FROM pairs")
            conn.commit()

    def save_pair(
        self, dill_id: int, doe_id: int, score: float, time_intersection: str
    ) -> None:
        """Save a matched pair."""
        with self.get_connection() as conn:
            conn.execute(
                """
                INSERT INTO pairs (dill_id, doe_id, score, time_intersection)
                VALUES (?, ?, ?, ?)
                """,
                (dill_id, doe_id, score, time_intersection),
            )
            conn.commit()

    def save_place(self, description: str) -> None:
        """Save a new place."""
        with self.get_connection() as conn:
            conn.execute(
                "INSERT INTO places (description) VALUES (?)",
                (description,),
            )
            conn.commit()

    def get_all_places(self) -> list[sqlite3.Row]:
        """Get all places."""
        with self.get_connection() as conn:
            cursor = conn.execute("SELECT * FROM places")
            return cursor.fetchall()

    def get_all_pairs(self) -> list[sqlite3.Row]:
        """Get all pairs."""
        with self.get_connection() as conn:
            cursor = conn.execute("SELECT * FROM pairs")
            return cursor.fetchall()

    def save_meeting(self, pair_id: int, place_id: int, time: str) -> int:
        """Save a meeting. Returns meeting_id."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "INSERT INTO meetings (pair_id, place_id, time) VALUES (?, ?, ?) RETURNING id",
                (pair_id, place_id, time),
            )
            row = cursor.fetchone()
            conn.commit()
            return row["id"]

    def get_user_by_id(self, user_id: int) -> Optional[sqlite3.Row]:
        """Get user by id (not telegram_id)."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT * FROM users WHERE id = ?",
                (user_id,),
            )
            return cursor.fetchone()

    def get_meeting_by_id(self, meeting_id: int) -> Optional[sqlite3.Row]:
        """Get meeting by id."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT * FROM meetings WHERE id = ?",
                (meeting_id,),
            )
            return cursor.fetchone()

    def get_pair_by_id(self, pair_id: int) -> Optional[sqlite3.Row]:
        """Get pair by id."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT * FROM pairs WHERE id = ?",
                (pair_id,),
            )
            return cursor.fetchone()

    def confirm_meeting(self, meeting_id: int, telegram_id: int) -> bool:
        """Confirm meeting attendance. Returns True if confirmed, False if user not in this meeting."""
        with self.get_connection() as conn:
            meeting = self.get_meeting_by_id(meeting_id)
            if not meeting:
                return False

            pair = self.get_pair_by_id(meeting["pair_id"])
            if not pair:
                return False

            dill = self.get_user_by_id(pair["dill_id"])
            doe = self.get_user_by_id(pair["doe_id"])

            if dill and dill["telegram_id"] == telegram_id:
                conn.execute(
                    "UPDATE meetings SET dill_confirmed = 1 WHERE id = ?",
                    (meeting_id,),
                )
                conn.commit()
                return True
            elif doe and doe["telegram_id"] == telegram_id:
                conn.execute(
                    "UPDATE meetings SET doe_confirmed = 1 WHERE id = ?",
                    (meeting_id,),
                )
                conn.commit()
                return True

        return False

    def get_partner_telegram_id(self, meeting_id: int, telegram_id: int) -> Optional[int]:
        """Get partner's telegram_id for a given meeting."""
        meeting = self.get_meeting_by_id(meeting_id)
        if not meeting:
            return None

        pair = self.get_pair_by_id(meeting["pair_id"])
        if not pair:
            return None

        dill = self.get_user_by_id(pair["dill_id"])
        doe = self.get_user_by_id(pair["doe_id"])

        if dill and dill["telegram_id"] == telegram_id:
            return doe["telegram_id"] if doe else None
        elif doe and doe["telegram_id"] == telegram_id:
            return dill["telegram_id"] if dill else None

        return None

    def get_partner_username(self, meeting_id: int, telegram_id: int) -> Optional[str]:
        """Get partner's username for a given meeting."""
        meeting = self.get_meeting_by_id(meeting_id)
        if not meeting:
            return None

        pair = self.get_pair_by_id(meeting["pair_id"])
        if not pair:
            return None

        dill = self.get_user_by_id(pair["dill_id"])
        doe = self.get_user_by_id(pair["doe_id"])

        if dill and dill["telegram_id"] == telegram_id:
            return doe["username"] if doe else None
        elif doe and doe["telegram_id"] == telegram_id:
            return dill["username"] if dill else None

        return None

    def get_user_username_by_telegram_id(self, telegram_id: int) -> Optional[str]:
        """Get username by telegram_id."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT username FROM users WHERE telegram_id = ?",
                (telegram_id,),
            )
            row = cursor.fetchone()
            return row["username"] if row else None

    def both_confirmed(self, meeting_id: int) -> bool:
        """Check if both participants confirmed the meeting."""
        meeting = self.get_meeting_by_id(meeting_id)
        if not meeting:
            return False
        return bool(meeting["dill_confirmed"]) and bool(meeting["doe_confirmed"])

    def cancel_meeting(self, meeting_id: int, telegram_id: int) -> bool:
        """Cancel meeting attendance. Returns True if cancelled, False if user not in this meeting."""
        with self.get_connection() as conn:
            meeting = self.get_meeting_by_id(meeting_id)
            if not meeting:
                return False

            pair = self.get_pair_by_id(meeting["pair_id"])
            if not pair:
                return False

            dill = self.get_user_by_id(pair["dill_id"])
            doe = self.get_user_by_id(pair["doe_id"])

            if dill and dill["telegram_id"] == telegram_id:
                conn.execute(
                    "UPDATE meetings SET dill_cancelled = 1 WHERE id = ?",
                    (meeting_id,),
                )
                conn.commit()
                return True
            elif doe and doe["telegram_id"] == telegram_id:
                conn.execute(
                    "UPDATE meetings SET doe_cancelled = 1 WHERE id = ?",
                    (meeting_id,),
                )
                conn.commit()
                return True

        return False
