import React from 'react';
import Button from '@mui/material/Button';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Stack, TextField, Typography, Box, List, ListItem, ListItemAvatar, Avatar, ListItemText, InputAdornment } from '@mui/material';
import SearchIcon from '@mui/icons-material;

const client = createDockerDesktopClient();

function useDockerDesktopClient() {
  return client;
}

type AppSummary = {
  name: string;
  description: string;
  repository: string;
  icon?: string;
};

export function App() {
  const [response, setResponse] = React.useState<string>();
  const [apps, setApps] = React.useState<AppSummary[]>([]);
  const [search, setSearch] = React.useState('');
  const ddClient = useDockerDesktopClient();

  const fetchAndDisplayResponse = async () => {
    const result = await ddClient.extension.vm?.service?.get('/hello');
    setResponse(JSON.stringify(result));
  };

  const fetchApps = async () => {
    const result = await ddClient.extension.vm?.service?.get('/apps');
    if (Array.isArray(result)) {
      setApps(result);
    }
  };

  React.useEffect(() => {
    fetchApps();
  }, []);

  const filteredApps = apps.filter(app =>
    app.name.toLowerCase().includes(search.toLowerCase()) ||
    app.description?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h3">Unraid Community Apps (Test in Docker Desktop)</Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
        Parcourez et testez les applications Unraid dans Docker Desktop avant de les lancer en production.
      </Typography>
      <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 4 }}>
        <Button variant="contained" onClick={fetchAndDisplayResponse}>
          Call backend
        </Button>
        <TextField
          label="Backend response"
          sx={{ width: 480 }}
          disabled
          multiline
          variant="outlined"
          minRows={5}
          value={response ?? ''}
        />
      </Stack>
      <TextField
        label="Rechercher une application"
        variant="outlined"
        sx={{ mt: 4, mb: 2, width: 400 }}
        value={search}
        onChange={e => setSearch(e.target.value)}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
      />
      <List>
        {filteredApps.map(app => (
          <ListItem key={app.name} alignItems="flex-start">
            <ListItemAvatar>
              <Avatar src={app.icon} alt={app.name} />
            </ListItemAvatar>
            <ListItemText
              primary={app.name}
              secondary={<>
                <Typography variant="body2" color="text.secondary">{app.description}</Typography>
                <Typography variant="caption" color="text.secondary">Image Docker: {app.repository}</Typography>
              </>}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );
}
