import { Box, Card, Flex, Text } from '@radix-ui/themes';
import type { Fractie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/fracties.schema';

export function FractieCard({ fractie }: { fractie: Fractie }) {
    return (
        <Box maxWidth="400px">
            <Card>
                <Flex gap="3" align="center" direction='column' width={'100%'}>
                    <Flex align="center" direction="row" justify='between' width="100%">
                        <Text as="div" size="2" weight="bold">
                            {fractie.naam_nl}
                        </Text>
                        <Text as="div" size="2" color="gray">
                            {fractie.aantal_stemmen}
                        </Text>
                        <Text as="div" size="2" color="gray">
                            {fractie.aantal_zetels}
                        </Text>
                    </Flex>
                    <Flex align="center" direction="row" justify='between' width="100%">
                        <Text as="div" size="2" color="gray">
                            {fractie.datum_actief?.toString()}
                        </Text>
                        <Text as="div" size="2" color="gray">
                            {fractie.datum_inactief?.toString()}
                        </Text>
                    </Flex>
                </Flex>
            </Card>
        </Box>
    )
}