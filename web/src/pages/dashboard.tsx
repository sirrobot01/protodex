import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card.tsx';
import { Button } from '@/components/ui/button.tsx';
import { Badge } from '@/components/ui/badge.tsx';
import { Package, Plus, Settings, Upload } from 'lucide-react';
import { useAuth } from '@/contexts/auth-context.tsx';

interface Package {
  id: string;
  name: string;
  description: string;
  version: string;
  downloads: number;
  stars: number;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export default function DashboardPage() {
  const [packages, setPackages] = useState<Package[]>([]);
  const [loading, setLoading] = useState(true);
  const { user, authenticatedFetch } = useAuth();


  useEffect(() => {
    fetchUserPackages();
  }, []);

  const fetchUserPackages = async () => {
    try {
      const response = await authenticatedFetch('/api/packages', {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setPackages(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch packages:', error);
    } finally {
      setLoading(false);
    }
  };


  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Welcome back, {user?.username}</h1>
          <p className="text-muted-foreground">
            Manage your protocol buffer packages.
          </p>
        </div>
        <div className="flex space-x-2">
          <Link to="/package/push">
            <Button variant="outline">
              <Upload className="h-4 w-4 mr-2" />
              Push Package
            </Button>
          </Link>
          <Link to="/package/create">
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Create Package
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Packages</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{packages.length}</div>
            <p className="text-xs text-muted-foreground">
              In your registry
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-2xl font-semibold">Your Packages ({packages.length})</h2>
        <PackageList packages={packages} />
      </div>
    </div>
  );
}

function PackageList({ 
  packages
}: { 
  packages: Package[]; 
}) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',

    });
  };

  if (packages.length === 0) {
    return (
      <div className="text-center py-12">
        <Package className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">No packages found</h3>
        <p className="text-muted-foreground mb-4">
          Create your first package to get started.
        </p>
        <Link to="/package/create">
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create Package
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {packages.map((pkg) => (
        <Card key={pkg.name} className="hover:shadow-md transition-shadow">
          <CardHeader>
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <CardTitle className="text-lg">
                  <Link 
                    to={`/package/${pkg.name}`}
                    className="hover:text-primary transition-colors"
                  >
                    {pkg.name}
                  </Link>
                </CardTitle>
                <CardDescription className="mt-1">
                  {pkg.version} • Created {formatDate(pkg.created_at)}
                </CardDescription>
              </div>
              
            </div>
            
            <CardDescription className="mt-2 line-clamp-2">
              {pkg.description}
            </CardDescription>
          </CardHeader>
          
          <CardContent>
            <div className="space-y-3">
              {pkg.tags.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {pkg.tags.slice(0, 3).map((tag) => (
                    <Badge key={tag} variant="outline" className="text-xs">
                      {tag}
                    </Badge>
                  ))}
                  {pkg.tags.length > 3 && (
                    <Badge variant="outline" className="text-xs">
                      +{pkg.tags.length - 3}
                    </Badge>
                  )}
                </div>
              )}

              <div className="flex items-center space-x-2 pt-2">
                <Link to={`/package/${pkg.name}`} className="flex-1">
                  <Button variant="outline" className="w-full">
                    <Settings className="h-4 w-4 mr-2" />
                    Manage
                  </Button>
                </Link>
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}