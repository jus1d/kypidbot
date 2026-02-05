from typing import Optional

from sqlalchemy import Boolean, Float, ForeignKey, Integer, String, select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

__all__ = ["Database", "UserModel", "PairModel", "PlaceModel", "MeetingModel"]


class Base(DeclarativeBase):
    pass


class UserModel(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    telegram_id: Mapped[int] = mapped_column(Integer, unique=True, nullable=False)
    username: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    first_name: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    last_name: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    is_bot: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    language_code: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    is_premium: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    sex: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    about: Mapped[str] = mapped_column(String, nullable=False, default="")
    state: Mapped[str] = mapped_column(String, nullable=False, default="start")
    time_ranges: Mapped[str] = mapped_column(String, nullable=False, default="000000")
    is_admin: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)


class PairModel(Base):
    __tablename__ = "pairs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    dill_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id"), nullable=False)
    doe_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id"), nullable=False)
    score: Mapped[float] = mapped_column(Float, nullable=False)
    time_intersection: Mapped[str] = mapped_column(String, nullable=False)
    is_fullmatch: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)


class PlaceModel(Base):
    __tablename__ = "places"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    description: Mapped[str] = mapped_column(String, nullable=False)


class MeetingModel(Base):
    __tablename__ = "meetings"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    pair_id: Mapped[int] = mapped_column(Integer, ForeignKey("pairs.id"), nullable=False)
    place_id: Mapped[int] = mapped_column(Integer, ForeignKey("places.id"), nullable=False)
    time: Mapped[str] = mapped_column(String, nullable=False)
    dill_confirmed: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    doe_confirmed: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    dill_cancelled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    doe_cancelled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)


class Database:
    def __init__(self, db_path: str = "kypidbot.db"):
        self.db_path = db_path
        self.engine = create_async_engine(f"sqlite+aiosqlite:///{db_path}", echo=False)
        self.async_session = async_sessionmaker(self.engine, class_=AsyncSession, expire_on_commit=False)

    async def init_db(self) -> None:
        """Initialize database schema."""
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

    async def save_user(
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
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.username = username
                user.first_name = first_name
                user.last_name = last_name
                user.is_bot = is_bot
                user.language_code = language_code
                user.is_premium = is_premium if is_premium is not None else False
            else:
                user = UserModel(
                    telegram_id=telegram_id,
                    username=username,
                    first_name=first_name,
                    last_name=last_name,
                    is_bot=is_bot,
                    language_code=language_code,
                    is_premium=is_premium if is_premium is not None else False,
                )
                session.add(user)

            await session.commit()

    async def get_time_ranges(self, telegram_id: int) -> str:
        """Get user's time ranges as binary string (e.g., '101010')."""
        async with self.async_session() as session:
            stmt = select(UserModel.time_ranges).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            time_ranges = result.scalar_one_or_none()
            return time_ranges if time_ranges else "000000"

    async def save_time_ranges(self, telegram_id: int, time_ranges: str) -> None:
        """Save user's time ranges as binary string."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.time_ranges = time_ranges
                await session.commit()

    async def get_user(self, telegram_id: int) -> Optional[UserModel]:
        """Get user by telegram_id."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()

    async def get_user_by_username(self, username: str) -> Optional[UserModel]:
        """Get user by username."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.username == username)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()

    async def set_user_sex(self, telegram_id: int, sex: str) -> None:
        """Set user's sex (male/female)."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.sex = sex
                await session.commit()

    async def set_user_about(self, telegram_id: int, about: str) -> None:
        """Set user's about/introduction text."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.about = about
                await session.commit()

    async def set_user_state(self, telegram_id: int, state: str) -> None:
        """Set user's current state in the flow."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.state = state
                await session.commit()

    async def get_user_state(self, telegram_id: int) -> str:
        """Get user's current state."""
        async with self.async_session() as session:
            stmt = select(UserModel.state).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            state = result.scalar_one_or_none()
            return state if state else "start"

    async def is_admin(self, telegram_id: int) -> bool:
        """Check if user is an admin."""
        async with self.async_session() as session:
            stmt = select(UserModel.is_admin).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            is_admin = result.scalar_one_or_none()
            return bool(is_admin) if is_admin is not None else False

    async def set_admin(self, telegram_id: int, is_admin: bool) -> None:
        """Set user's admin status."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            user = result.scalar_one_or_none()

            if user:
                user.is_admin = is_admin
                await session.commit()

    async def get_verified_users(self) -> list[UserModel]:
        """Get all users who completed verification."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.state == "completed")
            result = await session.execute(stmt)
            users = result.scalars().all()
            return list(users)

    async def clear_pairs(self) -> None:
        """Delete all existing pairs."""
        async with self.async_session() as session:
            await session.execute(select(PairModel))
            result = await session.execute(select(PairModel))
            pairs = result.scalars().all()
            for pair in pairs:
                await session.delete(pair)
            await session.commit()

    async def save_pair(
        self, dill_id: int, doe_id: int, score: float, time_intersection: str, is_fullmatch: bool = False
    ) -> None:
        """Save a matched pair."""
        async with self.async_session() as session:
            pair = PairModel(
                dill_id=dill_id,
                doe_id=doe_id,
                score=score,
                time_intersection=time_intersection,
                is_fullmatch=is_fullmatch,
            )
            session.add(pair)
            await session.commit()

    async def get_full_pairs(self) -> list[PairModel]:
        """Get all full pairs (full-match without time intersection)."""
        async with self.async_session() as session:
            stmt = select(PairModel).where(PairModel.is_fullmatch == True)  # noqa: E712
            result = await session.execute(stmt)
            return list(result.scalars().all())

    async def get_regular_pairs(self) -> list[PairModel]:
        """Get all regular pairs (with time intersection)."""
        async with self.async_session() as session:
            stmt = select(PairModel).where(PairModel.is_fullmatch == False)  # noqa: E712
            result = await session.execute(stmt)
            return list(result.scalars().all())

    async def save_place(self, description: str) -> None:
        """Save a new place."""
        async with self.async_session() as session:
            place = PlaceModel(description=description)
            session.add(place)
            await session.commit()

    async def get_all_places(self) -> list[PlaceModel]:
        """Get all places."""
        async with self.async_session() as session:
            stmt = select(PlaceModel)
            result = await session.execute(stmt)
            return list(result.scalars().all())

    async def get_all_pairs(self) -> list[PairModel]:
        """Get all pairs."""
        async with self.async_session() as session:
            stmt = select(PairModel)
            result = await session.execute(stmt)
            return list(result.scalars().all())

    async def save_meeting(self, pair_id: int, place_id: int, time: str) -> int:
        """Save a meeting. Returns meeting_id."""
        async with self.async_session() as session:
            meeting = MeetingModel(pair_id=pair_id, place_id=place_id, time=time)
            session.add(meeting)
            await session.commit()
            await session.refresh(meeting)
            return meeting.id

    async def get_user_by_id(self, user_id: int) -> Optional[UserModel]:
        """Get user by id (not telegram_id)."""
        async with self.async_session() as session:
            stmt = select(UserModel).where(UserModel.id == user_id)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()

    async def get_meeting_by_id(self, meeting_id: int) -> Optional[MeetingModel]:
        """Get meeting by id."""
        async with self.async_session() as session:
            stmt = select(MeetingModel).where(MeetingModel.id == meeting_id)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()

    async def get_pair_by_id(self, pair_id: int) -> Optional[PairModel]:
        """Get pair by id."""
        async with self.async_session() as session:
            stmt = select(PairModel).where(PairModel.id == pair_id)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()

    async def confirm_meeting(self, meeting_id: int, telegram_id: int) -> bool:
        """Confirm meeting attendance. Returns True if confirmed, False if user not in this meeting."""
        async with self.async_session() as session:
            meeting = await self.get_meeting_by_id(meeting_id)
            if not meeting:
                return False

            pair = await self.get_pair_by_id(meeting.pair_id)
            if not pair:
                return False

            dill = await self.get_user_by_id(pair.dill_id)
            doe = await self.get_user_by_id(pair.doe_id)

            stmt = select(MeetingModel).where(MeetingModel.id == meeting_id)
            result = await session.execute(stmt)
            meeting_to_update = result.scalar_one_or_none()

            if not meeting_to_update:
                return False

            if dill and dill.telegram_id == telegram_id:
                meeting_to_update.dill_confirmed = True
                await session.commit()
                return True
            elif doe and doe.telegram_id == telegram_id:
                meeting_to_update.doe_confirmed = True
                await session.commit()
                return True

        return False

    async def get_partner_telegram_id(self, meeting_id: int, telegram_id: int) -> Optional[int]:
        """Get partner's telegram_id for a given meeting."""
        meeting = await self.get_meeting_by_id(meeting_id)
        if not meeting:
            return None

        pair = await self.get_pair_by_id(meeting.pair_id)
        if not pair:
            return None

        dill = await self.get_user_by_id(pair.dill_id)
        doe = await self.get_user_by_id(pair.doe_id)

        if dill and dill.telegram_id == telegram_id:
            return doe.telegram_id if doe else None
        elif doe and doe.telegram_id == telegram_id:
            return dill.telegram_id if dill else None

        return None

    async def get_partner_username(self, meeting_id: int, telegram_id: int) -> Optional[str]:
        """Get partner's username for a given meeting."""
        meeting = await self.get_meeting_by_id(meeting_id)
        if not meeting:
            return None

        pair = await self.get_pair_by_id(meeting.pair_id)
        if not pair:
            return None

        dill = await self.get_user_by_id(pair.dill_id)
        doe = await self.get_user_by_id(pair.doe_id)

        if dill and dill.telegram_id == telegram_id:
            return doe.username if doe else None
        elif doe and doe.telegram_id == telegram_id:
            return dill.username if dill else None

        return None

    async def get_user_username_by_telegram_id(self, telegram_id: int) -> Optional[str]:
        """Get username by telegram_id."""
        async with self.async_session() as session:
            stmt = select(UserModel.username).where(UserModel.telegram_id == telegram_id)
            result = await session.execute(stmt)
            username = result.scalar_one_or_none()
            return username

    async def both_confirmed(self, meeting_id: int) -> bool:
        """Check if both participants confirmed the meeting."""
        meeting = await self.get_meeting_by_id(meeting_id)
        if not meeting:
            return False
        return meeting.dill_confirmed and meeting.doe_confirmed

    async def cancel_meeting(self, meeting_id: int, telegram_id: int) -> bool:
        """Cancel meeting attendance. Returns True if cancelled, False if user not in this meeting."""
        async with self.async_session() as session:
            meeting = await self.get_meeting_by_id(meeting_id)
            if not meeting:
                return False

            pair = await self.get_pair_by_id(meeting.pair_id)
            if not pair:
                return False

            dill = await self.get_user_by_id(pair.dill_id)
            doe = await self.get_user_by_id(pair.doe_id)

            stmt = select(MeetingModel).where(MeetingModel.id == meeting_id)
            result = await session.execute(stmt)
            meeting_to_update = result.scalar_one_or_none()

            if not meeting_to_update:
                return False

            if dill and dill.telegram_id == telegram_id:
                meeting_to_update.dill_cancelled = True
                await session.commit()
                return True
            elif doe and doe.telegram_id == telegram_id:
                meeting_to_update.doe_cancelled = True
                await session.commit()
                return True

        return False
