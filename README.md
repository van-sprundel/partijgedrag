# Partijgedrag - Dutch Voting Compass

A modern web application that helps Dutch citizens discover which political parties align best with their views by answering questions about real parliamentary motions and decisions.

## 🎯 Features

- **Interactive Voting Compass**: Answer 20 questions based on real parliamentary motions
- **Real Data**: Built using actual voting records from the Dutch Parliament (Tweede Kamer)
- **Party Matching**: Get detailed results showing which parties align with your views
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Fast & Modern**: Built with React, TypeScript, and modern web technologies

## 🏗️ Architecture

This is a full-stack application consisting of:

### Backend (`/backend`)
- **ORPC**: Type-safe API with automatic client generation
- **Prisma**: Database ORM with PostgreSQL
- **Express**: Web server framework
- **TypeScript**: Full type safety

### Frontend (`/frontend`)
- **React 18**: Modern React with hooks
- **TanStack Query**: Data fetching and caching
- **React Router**: Client-side routing
- **Tailwind CSS**: Utility-first styling
- **Vite**: Fast build tool and dev server

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- PostgreSQL 14+
- npm or yarn

### 1. Database Setup

```bash
# Create databases
createdb partijgedrag
```

### 2. Backend Setup

```bash
cd backend
npm install

# Copy environment file and configure
cp .env.example .env
# Edit .env with your database URL

# Generate Prisma client and push schema
npm run db:generate
npm run db:push

# Seed with sample data
npm run db:seed

# Start development server
npm run dev
```

The backend will be available at `http://localhost:3001`

### 3. Frontend Setup

```bash
cd frontend
npm install

# Copy environment file
cp .env.example .env

# Start development server
npm run dev
```

The frontend will be available at `http://localhost:3000`

## 📊 Data Model

The application works with several key entities:

- **Motion**: Parliamentary motions that were voted on
- **Party**: Political parties (VVD, D66, PvdA, etc.)
- **Politician**: Individual members of parliament
- **Vote**: How each politician voted on each motion
- **UserSession**: Stores user answers and calculated results

## 🔧 Development

### Backend Commands

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server
npm run db:generate  # Generate Prisma client
npm run db:push      # Push schema to database
npm run db:studio    # Open Prisma Studio
npm run db:seed      # Seed database with sample data
```

### Frontend Commands

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run preview  # Preview production build
npm run lint     # Run ESLint
```

### Environment Variables

#### Backend (`.env`)
```
DATABASE_URL="postgresql://username:password@localhost:5432/partijgedrag"
PORT=3001
CORS_ORIGIN=http://localhost:3000
```

#### Frontend (`.env`)
```
VITE_API_URL=http://localhost:3001
```

## 🌐 API Endpoints

The API is built with ORPC and provides type-safe endpoints:

### Motions
- `GET /api/motions/getAll` - Get paginated motions
- `GET /api/motions/getById` - Get specific motion
- `GET /api/motions/getForCompass` - Get motions for voting compass
- `GET /api/motions/getVotes` - Get votes for a motion

### Parties
- `GET /api/parties/getAll` - Get all parties
- `GET /api/parties/getById` - Get specific party
- `GET /api/parties/getWithVotes` - Get party with voting history

### Compass
- `POST /api/compass/submitAnswers` - Submit user answers
- `GET /api/compass/getResults` - Get results for session
- `GET /api/compass/getMotionDetails` - Get detailed motion info

## 🎨 UI Components

The frontend includes reusable UI components:

- `Button` - Styled button with variants and loading states
- `Card` - Content containers with header/footer
- `Progress` - Progress bars with different variants
- Custom hooks for API integration with TanStack Query

## 🔄 How It Works

1. **Question Selection**: The app selects 20 random motions from the database
2. **User Input**: Users answer agree/disagree/neutral for each motion
3. **Party Matching**: The backend compares user answers with actual party votes
4. **Results Calculation**: A matching algorithm calculates alignment percentages
5. **Results Display**: Users see which parties align best with their views

## 🚀 Deployment

### Backend Deployment

1. Set up PostgreSQL database
2. Configure environment variables
3. Run database migrations: `npm run db:push`
4. Seed data: `npm run db:seed`
5. Build: `npm run build`
6. Start: `npm run start`

### Frontend Deployment

1. Configure `VITE_API_URL` for production
2. Build: `npm run build`
3. Deploy `dist/` folder to static hosting (Vercel, Netlify, etc.)

## 🔗 Integration with ETL

This application is designed to work alongside the ETL system in the `/etl` directory:

1. ETL processes raw parliamentary data
2. Processed data can be imported into this app's database
3. The voting compass uses real parliamentary voting records

## 📝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Run tests and linting
5. Commit your changes: `git commit -am 'Add feature'`
6. Push to the branch: `git push origin feature-name`
7. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Data sourced from the official Dutch Parliament APIs
- Inspired by traditional voting compass tools like Stemwijzer
- Built with modern web technologies for optimal performance
