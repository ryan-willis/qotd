import { AppShell, Badge, Button, Flex, Group, Title } from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { notifications } from "@mantine/notifications";

const Header: React.FC<{
  latency: number;
  connected: boolean;
  roomId: string;
  onDisconnect: () => void;
}> = ({ onDisconnect, latency, connected, roomId }) => {
  const isMobile = useMediaQuery("(max-width: 36em)");
  return (
    <AppShell.Header>
      <Flex h="100%" justify="space-between">
        <Group h="100%" pl="md" justify="flex-start">
          <Title order={1}>QOTD</Title>
        </Group>
        <Group
          h="100%"
          pr="md"
          justify="flex-end"
          align="center"
          gap={isMobile ? "xs" : "md"}
        >
          {connected && (
            <>
              <Badge color={latency > -1 ? "pale-indigo" : "gray"}>
                {!isMobile && <>Ping: </>}
                {latency > -1 ? `${latency == 0 ? "<1" : latency}ms` : "N/A"}
              </Badge>
              {roomId && (
                <Badge
                  color="lime"
                  style={{
                    cursor: "pointer",
                  }}
                  onClick={() => {
                    const url = new URL(window.location.href);
                    url.searchParams.set("room", roomId);
                    const link = url.toString();
                    navigator.clipboard.writeText(link);
                    notifications.show({
                      title: "Room link copied to clipboard",
                      message: link,
                      autoClose: 3000,
                    });
                  }}
                >
                  {!isMobile && <>Room: </>}
                  {roomId}
                </Badge>
              )}
              {roomId && (
                <Button
                  color="gray"
                  size={isMobile ? "compact-sm" : "sm"}
                  onClick={() => onDisconnect()}
                >
                  Leave
                </Button>
              )}
            </>
          )}
        </Group>
      </Flex>
    </AppShell.Header>
  );
};

export default Header;
