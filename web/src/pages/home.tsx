import { Link, Navigate } from 'react-router-dom';
import { Button } from '@/components/ui/button.tsx';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card.tsx';
import { Package, Shield, Code } from 'lucide-react';
import { useAuth } from '@/contexts/auth-context.tsx';

export default function HomePage() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  if (user) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="space-y-16">
      <div className="text-center space-y-6 py-12">
        <h1 className="text-5xl font-bold bg-clip-text">
          Welcome to Protodex
        </h1>
        <p className="text-xl text-muted-foreground max-w-3xl mx-auto leading-relaxed">
          Your self-hosted registry for Protocol Buffer packages. Share and manage 
          your .proto files within your organization. Build better APIs with structured, 
          type-safe protocol definitions.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
          <Link to="/dashboard">
            <Button size="lg" className="px-8">
              View Dashboard
            </Button>
          </Link>
          <Link to="/register">
            <Button variant="outline" size="lg" className="px-8">
              Get Started
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <Card>
          <CardHeader>
            <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
              <Package className="h-6 w-6 text-primary" />
            </div>
            <CardTitle>Package Management</CardTitle>
            <CardDescription>
              Manage your organization's protocol buffer packages efficiently
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Organize and manage your .proto definitions in one place. Search by name, 
              tags, or functionality to quickly locate what you need.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
              <Shield className="h-6 w-6 text-primary" />
            </div>
            <CardTitle>Secure Hosting</CardTitle>
            <CardDescription>
              Host your protocol definitions securely with version control
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Keep your .proto files safe and accessible within your organization. 
              Secure hosting ensures your definitions are protected and available.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
              <Code className="h-6 w-6 text-primary" />
            </div>
            <CardTitle>Easy Integration</CardTitle>
            <CardDescription>
              Seamlessly integrate packages into your development workflow
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Simple CLI tools and APIs make it easy to publish, download, and 
              manage protocol buffer dependencies in your projects.
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Call to Action */}
      <div className="text-center space-y-6 py-12 bg-primary/5 rounded-lg">
        <h2 className="text-3xl font-bold">Ready to get started?</h2>
        <p className="text-muted-foreground max-w-2xl mx-auto">
          Start managing your organization's protocol buffer packages today. 
          Create an account to begin publishing and organizing your .proto files.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link to="/dashboard">
            <Button size="lg">
              View Dashboard
            </Button>
          </Link>
          <Link to="/register">
            <Button variant="outline" size="lg">
              Create Account
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}