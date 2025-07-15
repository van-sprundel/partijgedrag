import { Avatar, Card, Flex, Text } from "@radix-ui/themes"
import { Box } from "@radix-ui/themes/components/box"
import type { Motie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/moties.schema';





export default function MotionCard({ motie }: { motie: Motie }) {
    return (<Box maxWidth="240px">
        <Card>
            <Flex gap="3" align="center" direction='column' width={'100%'}>
                <Flex align="center" direction="row" justify='between' width="100%">
                    <Text as="div" size="2" weight="bold">
                        {motie.titel}
                    </Text>
                    <Text as="div" size="2" color="gray">
                        {motie.onderwerp}
                    </Text>
                    <Text as="div" size="2" color="gray">
                        {motie.vergaderjaar}
                    </Text>
                </Flex>
                <Flex align="center" direction="row" justify='between' width="100%">
                    <Text as="div" size="2" color="gray">
                        {motie.status}
                    </Text>
                    <Text as="div" size="2" color="gray">
                        {motie.gestartOp?.toString()}
                    </Text>
                </Flex>
            </Flex>
        </Card>
    </Box>)
}