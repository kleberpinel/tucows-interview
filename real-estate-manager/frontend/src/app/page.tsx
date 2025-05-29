import Link from 'next/link'

export default function Home() {
  return (
    <main className="container mx-auto px-4 py-8">
      <h1 className="text-4xl font-bold text-center mb-8">
        Real Estate Manager
      </h1>
      
      <div className="max-w-md mx-auto space-y-4">
        <Link 
          href="/login" 
          className="block w-full text-center bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        >
          Login
        </Link>
        
        <Link 
          href="/properties" 
          className="block w-full text-center bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
        >
          View Properties
        </Link>
      </div>
    </main>
  )
}